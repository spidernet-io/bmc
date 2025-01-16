package secret

import (
	"context"
	"github.com/spidernet-io/bmc/pkg/log"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/spidernet-io/bmc/pkg/agent/config"
	"github.com/spidernet-io/bmc/pkg/agent/hoststatus"
	"k8s.io/apimachinery/pkg/api/errors"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type SecretReconciler struct {
	client               client.Client
	config               *config.AgentConfig
	hostStatusController hoststatus.HostStatusController
}

// NewHostEndpointReconciler creates a new HostEndpoint reconciler
func NewSecretReconciler(mgr ctrl.Manager, config *config.AgentConfig, hostStatusController hoststatus.HostStatusController) (*SecretReconciler, error) {
	return &SecretReconciler{
		client:               mgr.GetClient(),
		config:               config,
		hostStatusController: hostStatusController,
	}, nil
}

// SetupWithManager sets up the controller with the Manager
func (r *SecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Secret{}).
		Complete(r)
}

// Reconcile handles the reconciliation of HostEndpoint objects
func (r *SecretReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := log.Logger.With(
		zap.String("secret", req.Name),
	)

	logger.Debugf("Reconciling Secret %s", req.Name)

	secret := &corev1.Secret{}
	if err := r.client.Get(ctx, req.NamespacedName, secret); err != nil {
		if errors.IsNotFound(err) {
			logger.Debugf("Secret not found, ignoring")
			return reconcile.Result{}, nil
		}
		logger.Error(err, "Failed to get Secret")
		return reconcile.Result{}, err
	}

	username := string(secret.Data["username"])
	password := string(secret.Data["password"])
	logger.Debugf("retrieved new secret data for %s/%s", secret.Namespace, secret.Name)
	r.hostStatusController.UpdateSecret(secret.Name, secret.Namespace, username, password)

	return reconcile.Result{}, nil
}
