package clusteragent

import (
	"context"
	"fmt"
	"os"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	bmcv1beta1 "github.com/spidernet-io/bmc/pkg/apis/bmc/v1beta1"
	"github.com/spidernet-io/bmc/pkg/controller/template"
)

// ClusterAgentReconciler reconciles a ClusterAgent object
type ClusterAgentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *ClusterAgentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the ClusterAgent instance
	clusterAgent := &bmcv1beta1.ClusterAgent{}
	err := r.Get(ctx, req.NamespacedName, clusterAgent)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Get the agent image from environment variable
	agentImage := os.Getenv("AGENT_IMAGE")
	if agentImage == "" {
		return ctrl.Result{}, fmt.Errorf("AGENT_IMAGE environment variable not set")
	}

	// Prepare template data
	data := &template.TemplateData{
		Name:              fmt.Sprintf("%s-agent", clusterAgent.Name),
		Namespace:         clusterAgent.Namespace,
		ClusterName:       clusterAgent.Spec.ClusterName,
		Image:            agentImage,
		Interface:        clusterAgent.Spec.Interface,
		Replicas:         clusterAgent.Spec.Replicas,
		ServiceAccountName: fmt.Sprintf("%s-agent-sa", clusterAgent.Name),
		RoleName:          fmt.Sprintf("%s-agent-role", clusterAgent.Name),
	}

	// Render all resources
	resources, err := template.RenderAllAgentResources(data)
	if err != nil {
		logger.Error(err, "Failed to render agent resources")
		return ctrl.Result{}, err
	}

	// Create or update all resources
	for kind, obj := range resources {
		// Set controller reference
		if err := controllerutil.SetControllerReference(clusterAgent, obj, r.Scheme); err != nil {
			logger.Error(err, "Failed to set controller reference", "kind", kind)
			return ctrl.Result{}, err
		}

		// Create or update the resource
		result, err := ctrl.CreateOrUpdate(ctx, r.Client, obj, func() error {
			// No need to update anything as we're using the rendered template
			return nil
		})
		if err != nil {
			logger.Error(err, "Failed to create or update resource", "kind", kind)
			return ctrl.Result{}, err
		}

		logger.Info("Resource reconciled", "kind", kind, "result", result)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterAgentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&bmcv1beta1.ClusterAgent{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&rbacv1.Role{}).
		Owns(&rbacv1.RoleBinding{}).
		Complete(r)
}
