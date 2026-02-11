package controller

import (
	"errors"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/seipan/cloudflared-dns-controller/pkg/cloudflare"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	testTargetName      = "cloudflared"
	testTargetNamespace = "cloudflared"
	testTargetKey       = "config.yaml"
	testTunnelID        = "test-tunnel-id"

	configYAML = `tunnel: test-tunnel-id
credentials-file: /etc/cloudflared/creds/credentials.json
metrics: 0.0.0.0:2000
no-autoupdate: true
ingress:
  - hostname: app.example.com
    service: http://traefik.traefik.svc.cluster.local:80
  - hostname: api.example.com
    service: http://traefik.traefik.svc.cluster.local:80
  - service: http_status:404
`
)

func tunnelTarget() string {
	return testTunnelID + ".cfargotunnel.com"
}

func newReconciler(fake *fakeCloudflareClient) *CloudflaredDNSReconciler {
	return &CloudflaredDNSReconciler{
		Client:          k8sClient,
		Scheme:          scheme.Scheme,
		Cloudflare:      fake,
		TargetName:      testTargetName,
		TargetNamespace: testTargetNamespace,
		TargetKey:       testTargetKey,
	}
}

func newConfigMap(data map[string]string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testTargetName,
			Namespace: testTargetNamespace,
		},
		Data: data,
	}
}

var _ = Describe("CloudflaredDNS Controller", func() {
	var (
		fakeCF     *fakeCloudflareClient
		reconciler *CloudflaredDNSReconciler
		req        ctrl.Request
	)

	BeforeEach(func() {
		ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: testTargetNamespace}}
		_ = k8sClient.Create(ctx, ns)

		fakeCF = &fakeCloudflareClient{}
		reconciler = newReconciler(fakeCF)
		req = ctrl.Request{
			NamespacedName: types.NamespacedName{
				Name:      testTargetName,
				Namespace: testTargetNamespace,
			},
		}
	})

	AfterEach(func() {
		cm := &corev1.ConfigMap{}
		if err := k8sClient.Get(ctx, req.NamespacedName, cm); err == nil {
			controllerutil.RemoveFinalizer(cm, finalizerName)
			_ = k8sClient.Update(ctx, cm)
			_ = k8sClient.Delete(ctx, cm)
		}
	})

	Context("Normal reconciliation", func() {
		It("should add finalizer and create DNS records for a new ConfigMap", func() {
			cm := newConfigMap(map[string]string{testTargetKey: configYAML})
			Expect(k8sClient.Create(ctx, cm)).To(Succeed())

			result, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(Equal(5 * time.Minute))

			By("verifying finalizer is added")
			Expect(k8sClient.Get(ctx, req.NamespacedName, cm)).To(Succeed())
			Expect(controllerutil.ContainsFinalizer(cm, finalizerName)).To(BeTrue())

			By("verifying DNS records are created")
			Expect(fakeCF.createdRecords).To(HaveLen(2))
			Expect(fakeCF.createdRecords[0].Name).To(Equal("app.example.com"))
			Expect(fakeCF.createdRecords[1].Name).To(Equal("api.example.com"))
			Expect(fakeCF.createdRecords[0].Content).To(Equal(tunnelTarget()))
			Expect(fakeCF.deletedIDs).To(BeEmpty())
		})

		It("should only create DNS records for new hostnames", func() {
			cm := newConfigMap(map[string]string{testTargetKey: configYAML})
			Expect(k8sClient.Create(ctx, cm)).To(Succeed())

			fakeCF.records = []cloudflare.DNSRecord{
				{ID: "rec-1", Name: "app.example.com", Type: "CNAME", Content: tunnelTarget()},
			}

			result, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(Equal(5 * time.Minute))

			Expect(fakeCF.createdRecords).To(HaveLen(1))
			Expect(fakeCF.createdRecords[0].Name).To(Equal("api.example.com"))
			Expect(fakeCF.deletedIDs).To(BeEmpty())
		})

		It("should delete DNS records for removed hostnames", func() {
			cm := newConfigMap(map[string]string{testTargetKey: configYAML})
			Expect(k8sClient.Create(ctx, cm)).To(Succeed())

			fakeCF.records = []cloudflare.DNSRecord{
				{ID: "rec-1", Name: "app.example.com", Type: "CNAME", Content: tunnelTarget()},
				{ID: "rec-2", Name: "api.example.com", Type: "CNAME", Content: tunnelTarget()},
				{ID: "rec-3", Name: "removed.example.com", Type: "CNAME", Content: tunnelTarget()},
			}

			result, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(Equal(5 * time.Minute))

			Expect(fakeCF.createdRecords).To(BeEmpty())
			Expect(fakeCF.deletedIDs).To(ConsistOf("rec-3"))
		})

		It("should not create or delete when no changes are needed", func() {
			cm := newConfigMap(map[string]string{testTargetKey: configYAML})
			Expect(k8sClient.Create(ctx, cm)).To(Succeed())

			fakeCF.records = []cloudflare.DNSRecord{
				{ID: "rec-1", Name: "app.example.com", Type: "CNAME", Content: tunnelTarget()},
				{ID: "rec-2", Name: "api.example.com", Type: "CNAME", Content: tunnelTarget()},
			}

			result, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(Equal(5 * time.Minute))

			Expect(fakeCF.createdRecords).To(BeEmpty())
			Expect(fakeCF.deletedIDs).To(BeEmpty())
		})

		It("should delete all DNS records and remove finalizer on ConfigMap deletion", func() {
			cm := newConfigMap(map[string]string{testTargetKey: configYAML})
			Expect(k8sClient.Create(ctx, cm)).To(Succeed())

			By("first reconcile to add finalizer")
			_, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())

			By("deleting ConfigMap with existing DNS records")
			fakeCF.createdRecords = nil
			fakeCF.records = []cloudflare.DNSRecord{
				{ID: "rec-1", Name: "app.example.com", Type: "CNAME", Content: tunnelTarget()},
				{ID: "rec-2", Name: "api.example.com", Type: "CNAME", Content: tunnelTarget()},
			}
			Expect(k8sClient.Delete(ctx, cm)).To(Succeed())

			By("second reconcile triggers deletion handling")
			result, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(BeZero())

			Expect(fakeCF.deletedIDs).To(ConsistOf("rec-1", "rec-2"))

			By("verifying ConfigMap is fully deleted")
			err = k8sClient.Get(ctx, req.NamespacedName, &corev1.ConfigMap{})
			Expect(client.IgnoreNotFound(err)).NotTo(HaveOccurred())
		})

		It("should do nothing when target key is missing", func() {
			cm := newConfigMap(map[string]string{"other.yaml": "dummy"})
			Expect(k8sClient.Create(ctx, cm)).To(Succeed())

			result, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(BeZero())

			Expect(fakeCF.createdRecords).To(BeEmpty())
			Expect(fakeCF.deletedIDs).To(BeEmpty())
		})
	})

	Context("Error handling", func() {
		It("should return error when ListDNSRecords fails", func() {
			cm := newConfigMap(map[string]string{testTargetKey: configYAML})
			Expect(k8sClient.Create(ctx, cm)).To(Succeed())

			fakeCF.listErr = errors.New("cloudflare api error")

			_, err := reconciler.Reconcile(ctx, req)
			Expect(err).To(MatchError(ContainSubstring("cloudflare api error")))
		})

		It("should return error when CreateDNSRecord fails", func() {
			cm := newConfigMap(map[string]string{testTargetKey: configYAML})
			Expect(k8sClient.Create(ctx, cm)).To(Succeed())

			fakeCF.createErr = errors.New("create failed")

			_, err := reconciler.Reconcile(ctx, req)
			Expect(err).To(MatchError(ContainSubstring("create failed")))
		})

		It("should keep finalizer when DeleteDNSRecord fails during deletion", func() {
			cm := newConfigMap(map[string]string{testTargetKey: configYAML})
			Expect(k8sClient.Create(ctx, cm)).To(Succeed())

			By("first reconcile to add finalizer")
			_, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())

			fakeCF.records = []cloudflare.DNSRecord{
				{ID: "rec-1", Name: "app.example.com", Type: "CNAME", Content: tunnelTarget()},
			}
			fakeCF.deleteErr = errors.New("delete failed")

			Expect(k8sClient.Delete(ctx, cm)).To(Succeed())

			_, err = reconciler.Reconcile(ctx, req)
			Expect(err).To(MatchError(ContainSubstring("delete failed")))

			By("verifying finalizer is still present")
			Expect(k8sClient.Get(ctx, req.NamespacedName, cm)).To(Succeed())
			Expect(controllerutil.ContainsFinalizer(cm, finalizerName)).To(BeTrue())
		})
	})
})
