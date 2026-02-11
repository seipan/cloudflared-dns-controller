package controller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("MarkdownView Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "sample"
		const testNamespace = "test"
		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: testNamespace,
		}

		BeforeEach(func() {
			By("creating the namespace for the test")
			ns := &corev1.Namespace{}
			ns.Name = testNamespace
			err := k8sClient.Create(context.Background(), ns)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.

		})

		It("should successfully reconcile the resource", func() {

		})
	})
})
