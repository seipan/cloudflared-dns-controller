package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/seipan/cloudflared-dns-controller/pkg/cloudflare"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type CloudflaredDNSReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Cloudflare cloudflare.Client

	TargetName      string // ex "cloudflared"
	TargetNamespace string // ex "cloudflared"
	TargetKey       string // ex "config.yaml"
}

func (r *CloudflaredDNSReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
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
