package clusteragent

import (
	"context"
	"fmt"
	"os"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// Reconcile is part of the main kubernetes reconciliation loop
func (r *ClusterAgentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues(
		"controller", "clusteragent",
		"namespace", req.Namespace,
		"name", req.Name,
	)

	controllerNS := os.Getenv("POD_NAMESPACE")
	if controllerNS == "" {
		logger.Error(nil, "POD_NAMESPACE environment variable not set")
		return ctrl.Result{}, fmt.Errorf("POD_NAMESPACE environment variable not set")
	}
	
	// Fetch the ClusterAgent instance
	clusterAgent := &bmcv1beta1.ClusterAgent{}
	err := r.Get(ctx, req.NamespacedName, clusterAgent)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("ClusterAgent resource not found, initiating cleanup")

			// Clean up resources
			objects := []client.Object{
				&appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      fmt.Sprintf("agent-%s", req.Name),
						Namespace: controllerNS,
					},
				},
				&corev1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{
						Name:      fmt.Sprintf("agent-%s", req.Name),
						Namespace: controllerNS,
					},
				},
				&rbacv1.Role{
					ObjectMeta: metav1.ObjectMeta{
						Name:      fmt.Sprintf("agent-%s", req.Name),
						Namespace: controllerNS,
					},
				},
				&rbacv1.RoleBinding{
					ObjectMeta: metav1.ObjectMeta{
						Name:      fmt.Sprintf("agent-%s", req.Name),
						Namespace: controllerNS,
					},
				},
			}

			for _, obj := range objects {
				if err := r.Delete(ctx, obj); err != nil && !errors.IsNotFound(err) {
					logger.Error(err, "Failed to delete resource",
						"kind", obj.GetObjectKind().GroupVersionKind().Kind,
						"name", obj.GetName(),
						"namespace", obj.GetNamespace())
				} else {
					logger.Info("Successfully deleted resource",
						"kind", obj.GetObjectKind().GroupVersionKind().Kind,
						"name", obj.GetName(),
						"namespace", obj.GetNamespace())
				}
			}

			logger.Info("Resource cleanup completed")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get ClusterAgent resource")
		return ctrl.Result{}, err
	}

	logger.Info("Starting reconciliation for ClusterAgent",
		"namespace", clusterAgent.Namespace,
		"name", clusterAgent.Name,
		"clusterName", clusterAgent.Spec.ClusterName)

	// Get the agent image from environment variable or spec
	agentImage := clusterAgent.Spec.Image
	if agentImage == "" {
		agentImage = os.Getenv("AGENT_IMAGE")
		if agentImage == "" {
			err := fmt.Errorf("neither spec.image nor AGENT_IMAGE environment variable is set")
			logger.Error(err, "No agent image configuration available")
			return ctrl.Result{}, err
		}
	}

	logger.Info("Retrieved agent image configuration",
		"image", agentImage)

	// Prepare template data
	name := fmt.Sprintf("agent-%s", clusterAgent.Spec.ClusterName)
	replicas := int32(1)
	if clusterAgent.Spec.Replicas != nil {
		replicas = *clusterAgent.Spec.Replicas
	}

	data := &template.TemplateData{
		Name:              name,
		Namespace:         controllerNS,
		ClusterName:       clusterAgent.Spec.ClusterName,
		Image:            agentImage,
		Replicas:         replicas,
		ServiceAccountName: name,
		RoleName:         name,
		UnderlayInterface: clusterAgent.Spec.UnderlayInterface,
	}

	logger.Info("Prepared template data for resource generation",
		"name", data.Name,
		"namespace", data.Namespace,
		"clusterName", data.ClusterName,
		"image", data.Image,
		"replicas", data.Replicas,
		"serviceAccount", data.ServiceAccountName,
		"roleName", data.RoleName,
		"underlayInterface", data.UnderlayInterface)

	// Render all resources
	resources, err := template.RenderAllAgentResources(data)
	if err != nil {
		logger.Error(err, "Failed to render agent resources")
		return ctrl.Result{}, err
	}

	logger.Info("Successfully rendered resources",
		"resourceCount", len(resources))

	// Create or update all resources
	for kind, obj := range resources {
		logger.Info("Processing resource",
			"kind", kind,
			"name", obj.GetName(),
			"namespace", obj.GetNamespace())

		// Set controller reference
		if err := controllerutil.SetControllerReference(clusterAgent, obj, r.Scheme); err != nil {
			logger.Error(err, "Failed to set controller reference",
				"kind", kind,
				"name", obj.GetName())
			return ctrl.Result{}, err
		}

		// Create or update the resource
		result, err := ctrl.CreateOrUpdate(ctx, r.Client, obj, func() error {
			return nil
		})

		if err != nil {
			logger.Error(err, "Failed to create/update resource",
				"kind", kind,
				"name", obj.GetName())
			return ctrl.Result{}, err
		}

		logger.Info("Resource operation completed",
			"kind", kind,
			"name", obj.GetName(),
			"operation", result)
	}

	// Check deployment status
	deployment := &appsv1.Deployment{}
	err = r.Get(ctx, client.ObjectKey{
		Namespace: controllerNS,
		Name:      name,
	}, deployment)

	if err != nil {
		logger.Error(err, "Failed to get deployment")
		return ctrl.Result{}, err
	}

	// Update ClusterAgent status.Ready field
	ready := deployment.Status.ReadyReplicas == deployment.Status.Replicas &&
		deployment.Status.UpdatedReplicas == deployment.Status.Replicas

	if ready != clusterAgent.Status.Ready {
		clusterAgent.Status.Ready = ready
		if err := r.Status().Update(ctx, clusterAgent); err != nil {
			logger.Error(err, "Failed to update ClusterAgent status")
			return ctrl.Result{}, err
		}
		logger.Info("Updated ClusterAgent status",
			"ready", ready)
	}

	// Requeue to periodically check deployment status
	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterAgentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&bmcv1beta1.ClusterAgent{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
