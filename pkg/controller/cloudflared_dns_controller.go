package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/go-logr/logr"
	"github.com/seipan/cloudflared-dns-controller/pkg/cloudflare"
	"github.com/seipan/cloudflared-dns-controller/pkg/config"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const finalizerName = "cloudflared-dns-controller.seipan.github.io/finalizer"

type CloudflaredDNSReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Cloudflare cloudflare.Client

	TargetName      string // ex "cloudflared"
	TargetNamespace string // ex "cloudflared"
	TargetKey       string // ex "config.yaml"
}

func (r *CloudflaredDNSReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	cm := &corev1.ConfigMap{}
	if err := r.Get(ctx, req.NamespacedName, cm); err != nil {
		log.Error(err, "unable to fetch ConfigMap")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if cm.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, log, cm)
	}

	if !controllerutil.ContainsFinalizer(cm, finalizerName) {
		controllerutil.AddFinalizer(cm, finalizerName)
		if err := r.Update(ctx, cm); err != nil {
			log.Error(err, "unable to add finalizer to ConfigMap")
			return ctrl.Result{}, err
		}
		log.Info("Finalizer added to ConfigMap")
	}

	return ctrl.Result{}, nil
}

func (r *CloudflaredDNSReconciler) handleDeletion(ctx context.Context, log logr.Logger, cm *corev1.ConfigMap) (ctrl.Result, error) {
	if !controllerutil.ContainsFinalizer(cm, finalizerName) {
		return ctrl.Result{}, nil
	}
	data, ok := cm.Data[r.TargetKey]
	if ok {
		cfg, err := config.Parse(data)
		if err != nil {
			return ctrl.Result{}, err
		}
		tunnelID := cfg.Tunnel

		for _, hostname := range cfg.Hostnames() {
			if r.Cloudflare.IsTunnelRecord(cloudflare.DNSRecord{Name: hostname}, tunnelID) {
				log.Info("Deleting DNS record", "hostname", hostname)
				if err := r.Cloudflare.DeleteDNSRecord(ctx, hostname); err != nil {
					return ctrl.Result{}, err
				}
			}
		}
	}

	controllerutil.RemoveFinalizer(cm, finalizerName)
	if err := r.Update(ctx, cm); err != nil {
		log.Error(err, "unable to remove finalizer from ConfigMap")
		return ctrl.Result{}, err
	}
	log.Info("Finalizer removed from ConfigMap")
	return ctrl.Result{}, nil
}

func (r *CloudflaredDNSReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.ConfigMap{}).
		WithEventFilter(predicate.NewPredicateFuncs(func(obj client.Object) bool {
			return obj.GetName() == r.TargetName &&
				obj.GetNamespace() == r.TargetNamespace
		})).
		Complete(r)
}
