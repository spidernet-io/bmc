package deployment

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	bmcv1beta1 "github.com/spidernet-io/bmc/pkg/apis/bmc/v1beta1"
	"github.com/spidernet-io/bmc/pkg/constants"
	"github.com/spidernet-io/bmc/pkg/controller/template"
)

// Manager handles deployment operations
type Manager struct {
	client      client.Client
	scheme      *runtime.Scheme
	tmplManager *template.Manager
}

// NewManager creates a new deployment manager
func NewManager(client client.Client, scheme *runtime.Scheme) *Manager {
	return &Manager{
		client:      client,
		scheme:      scheme,
		tmplManager: template.NewManager("/etc/bmc/templates"),
	}
}

// CreateOrUpdate creates or updates the agent deployment
func (m *Manager) CreateOrUpdate(ctx context.Context, agent *bmcv1beta1.ClusterAgent, namespace, defaultImage string) error {
	// Prepare template data
	data := map[string]interface{}{
		"NAME":         fmt.Sprintf("%s-agent", agent.Name),
		"NAMESPACE":    namespace,
		"CLUSTER_NAME": agent.Spec.ClusterName,
		"IMAGE":        agent.Spec.Image,
		"INTERFACE":    agent.Spec.Interface,
		"REPLICAS":     agent.Spec.Replicas,
	}

	// Render deployment from template
	obj, err := m.tmplManager.RenderYAML("agent-deployment.yaml", data)
	if err != nil {
		return fmt.Errorf("failed to render deployment template: %v", err)
	}

	// Convert to deployment
	deployment := &appsv1.Deployment{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, deployment); err != nil {
		return fmt.Errorf("failed to convert unstructured to deployment: %v", err)
	}

	// Set owner reference
	if err := controllerutil.SetControllerReference(agent, deployment, m.scheme); err != nil {
		return err
	}

	// Create or update the deployment
	existing := &appsv1.Deployment{}
	err = m.client.Get(ctx, types.NamespacedName{
		Name:      deployment.Name,
		Namespace: deployment.Namespace,
	}, existing)

	if err != nil {
		if errors.IsNotFound(err) {
			// Create new deployment
			if err := m.client.Create(ctx, deployment); err != nil {
				return fmt.Errorf("failed to create deployment: %v", err)
			}
		} else {
			return fmt.Errorf("failed to get deployment: %v", err)
		}
	} else {
		// Update existing deployment
		existing.Spec = deployment.Spec
		if err := m.client.Update(ctx, existing); err != nil {
			return fmt.Errorf("failed to update deployment: %v", err)
		}
	}

	return nil
}

// IsReady checks if the deployment is ready
func (m *Manager) IsReady(deployment *appsv1.Deployment) bool {
	return deployment.Status.ReadyReplicas == *deployment.Spec.Replicas
}

// ShouldUpdate checks if the deployment needs to be updated
func (m *Manager) ShouldUpdate(deployment *appsv1.Deployment, agent *bmcv1beta1.ClusterAgent, defaultImage string) bool {
	// Check if replicas need update
	currentReplicas := *deployment.Spec.Replicas
	desiredReplicas := agent.Spec.Replicas
	if desiredReplicas == 0 {
		desiredReplicas = 1
	}
	if currentReplicas != desiredReplicas {
		return true
	}

	// Check if image needs update
	currentImage := deployment.Spec.Template.Spec.Containers[0].Image
	desiredImage := defaultImage
	if agent.Spec.Image != "" {
		desiredImage = agent.Spec.Image
	}
	if currentImage != desiredImage {
		return true
	}

	// Check if interface annotation needs update
	currentInterface := deployment.Spec.Template.ObjectMeta.Annotations[constants.NetworkAnnotationKey]
	if currentInterface != agent.Spec.Interface {
		return true
	}

	return false
}

// Delete deletes the agent deployment
func (m *Manager) Delete(ctx context.Context, name, namespace string) error {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	if err := m.client.Delete(ctx, deployment); err != nil && !errors.IsNotFound(err) {
		return err
	}
	return nil
}
