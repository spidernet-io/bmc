package clusteragent

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"sync"
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

	bmcv1beta1 "github.com/spidernet-io/bmc/pkg/apis/bmc/v1beta1"
	"github.com/spidernet-io/bmc/pkg/log"
	"github.com/spidernet-io/bmc/pkg/controller/template"
	"go.uber.org/zap"
)

// ClusterAgentReconciler reconciles a ClusterAgent object
type ClusterAgentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	cache  sync.Map // Store ClusterAgent instances in local cache
}

var GlobalControllerNS string

func init() {
	GlobalControllerNS = os.Getenv("POD_NAMESPACE")
	if GlobalControllerNS == "" {
		panic("POD_NAMESPACE environment variable not set")
	}
}

// getFromCache safely retrieves a ClusterAgent from the cache
// Returns a deep copy of the cached object to prevent concurrent modifications
func (r *ClusterAgentReconciler) getFromCache(name string) *bmcv1beta1.ClusterAgent {
	if value, ok := r.cache.Load(name); ok {
		if cached, ok := value.(*bmcv1beta1.ClusterAgent); ok {
			return cached.DeepCopy() // Return a deep copy to prevent concurrent modifications
		}
	}
	return nil
}

// storeInCache safely stores a ClusterAgent in the cache
// If agent is nil, removes the entry from cache
// Stores a deep copy to prevent external modifications affecting the cache
func (r *ClusterAgentReconciler) storeInCache(name string, agent *bmcv1beta1.ClusterAgent) {
	if agent == nil {
		r.cache.Delete(name)
		return
	}
	r.cache.Store(name, agent.DeepCopy()) // Store a deep copy to prevent external modifications
}

// hasSpecChanged checks if the spec has been modified
// Returns true if old is nil or if specs are different
func (r *ClusterAgentReconciler) hasSpecChanged(old, new *bmcv1beta1.ClusterAgent) bool {
	if old == nil {
		return true
	}
	return !reflect.DeepEqual(old.Spec, new.Spec)
}

// cleanupResources removes all resources associated with a ClusterAgent
// This includes the deployment, service account, role, and role binding
// Also removes the instance from the local cache
func (r *ClusterAgentReconciler) cleanupResources(ctx context.Context, name string, logger *zap.SugaredLogger) error {
	// First remove from cache
	r.cache.Delete(name)

	objects := []client.Object{
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("agent-%s", name),
				Namespace: GlobalControllerNS,
			},
		},
		&corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("agent-%s", name),
				Namespace: GlobalControllerNS,
			},
		},
		&rbacv1.Role{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("agent-%s", name),
				Namespace: GlobalControllerNS,
			},
		},
		&rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("agent-%s", name),
				Namespace: GlobalControllerNS,
			},
		},
	}

	for _, obj := range objects {
		if err := r.Delete(ctx, obj); err != nil && !errors.IsNotFound(err) {
			logger.Errorf("Failed to delete resource %s/%s: %v", 
				obj.GetNamespace(), obj.GetName(), err)
			return err
		}
		logger.Infof("Successfully deleted resource %s/%s", 
			obj.GetNamespace(), obj.GetName())
	}

	return nil
}

// createOrUpdateResources creates or updates all resources for a ClusterAgent
// This includes the deployment, service account, role, and role binding
func (r *ClusterAgentReconciler) createOrUpdateResources(ctx context.Context, clusterAgent *bmcv1beta1.ClusterAgent, logger *zap.SugaredLogger) error {
	// Get agent image from spec or environment variable
	agentImage := clusterAgent.Spec.AgentYaml.Image
	if agentImage == "" {
		agentImage = os.Getenv("AGENT_IMAGE")
		if agentImage == "" {
			return fmt.Errorf("neither spec.image nor AGENT_IMAGE environment variable is set")
		}
	}

	logger.Infof("Using agent image: %s", agentImage)

	// Prepare template data
	name := fmt.Sprintf("agent-%s", clusterAgent.Name)
	var replicas int32 = 1
	if clusterAgent.Spec.AgentYaml.Replicas != nil {
		replicas = *clusterAgent.Spec.AgentYaml.Replicas
	}

	data := &template.TemplateData{
		Name:               name,
		Namespace:          GlobalControllerNS,
		ClusterName:        clusterAgent.Name,
		Image:             agentImage,
		Replicas:          replicas,
		ServiceAccountName: name,
		RoleName:          name,
		UnderlayInterface:  clusterAgent.Spec.AgentYaml.UnderlayInterface,
		NodeAffinity:      clusterAgent.Spec.AgentYaml.NodeAffinity,
		NodeName:          clusterAgent.Spec.AgentYaml.NodeName,
	}

	// Render resources from template
	resources, err := template.RenderAllAgentResources(data)
	if err != nil {
		return fmt.Errorf("failed to render agent resources: %v", err)
	}

	// Create or update each resource
	for kind, obj := range resources {
		if err := controllerutil.SetControllerReference(clusterAgent, obj, r.Scheme); err != nil {
			return fmt.Errorf("failed to set controller reference for %s %s/%s: %v",
				kind, obj.GetNamespace(), obj.GetName(), err)
		}

		result, err := ctrl.CreateOrUpdate(ctx, r.Client, obj, func() error {
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to create/update resource %s %s/%s: %v",
				kind, obj.GetNamespace(), obj.GetName(), err)
		}

		logger.Infof("Resource operation completed: %s %s/%s (%s)",
			kind, obj.GetNamespace(), obj.GetName(), result)
	}

	return nil
}

// updateStatus updates the ClusterAgent status based on the deployment state
// Returns:
// - error: any error that occurred during the update
// - bool: true if status was changed, false otherwise
func (r *ClusterAgentReconciler) updateStatus(ctx context.Context, clusterAgent *bmcv1beta1.ClusterAgent, logger *zap.SugaredLogger) (error, bool) {
	// Default to not ready
	ready := false
	deployment := &appsv1.Deployment{}
	err := r.Get(ctx, client.ObjectKey{
		Namespace: GlobalControllerNS,
		Name:      fmt.Sprintf("agent-%s", clusterAgent.Name),
	}, deployment)

	// Only set ready=true if:
	// 1. Deployment exists
	// 2. All replicas are ready
	// 3. All replicas are updated
	// 4. No error conditions present
	if err == nil &&
		deployment.Status.Replicas > 0 &&
		deployment.Status.ReadyReplicas == deployment.Status.Replicas &&
		deployment.Status.UpdatedReplicas == deployment.Status.Replicas &&
		deployment.Status.AvailableReplicas == deployment.Status.Replicas &&
		len(deployment.Status.Conditions) > 0 {
		// Check if deployment is truly available
		for _, condition := range deployment.Status.Conditions {
			if condition.Type == appsv1.DeploymentAvailable &&
				condition.Status == corev1.ConditionTrue {
				ready = true
				break
			}
		}
	}

	// Update status if it has changed
	if ready != clusterAgent.Status.Ready {
		clusterAgent.Status.Ready = ready
		if err := r.Status().Update(ctx, clusterAgent); err != nil {
			return fmt.Errorf("failed to update status: %v", err), false
		}
		logger.Infof("Updated ClusterAgent status: ready=%v", ready)
		return nil, true
	}

	return nil, false
}

// Reconcile is part of the main kubernetes reconciliation loop
// It implements the reconciliation logic for ClusterAgent resources
func (r *ClusterAgentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.Logger.With(
		zap.String("reconcile", "clusteragent"),
		zap.String("name", req.Name),
	)

	// Get the ClusterAgent instance
	clusterAgent := &bmcv1beta1.ClusterAgent{}
	err := r.Get(ctx, req.NamespacedName, clusterAgent)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Infof("ClusterAgent resource not found: %s, initiating cleanup", req.Name)
			if err := r.cleanupResources(ctx, req.Name, logger); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Get previous instance from cache
	oldClusterAgent := r.getFromCache(req.Name)

	// Check if spec has changed
	if r.hasSpecChanged(oldClusterAgent, clusterAgent) {
		logger.Infof("Spec changed, updating resources")
		if err := r.createOrUpdateResources(ctx, clusterAgent, logger); err != nil {
			return ctrl.Result{}, err
		}
		// Update cache with new instance
		r.storeInCache(req.Name, clusterAgent)
		logger.Infof("succeeded to create k8s resource for agentCluster %s", clusterAgent.Name)
	} else {
		logger.Debugf("Spec unchanged, skipping resource update")
	}

	// Update status regardless of spec changes
	if err, statusChanged := r.updateStatus(ctx, clusterAgent, logger); err != nil {
		return ctrl.Result{}, err
	} else if statusChanged {
		logger.Infof("Status changed for agentCluster %s", clusterAgent.Name)
	}

	return ctrl.Result{RequeueAfter: 60 * time.Second}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterAgentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&bmcv1beta1.ClusterAgent{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
