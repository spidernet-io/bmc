package deployment

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	bmcv1beta1 "github.com/spidernet-io/bmc/pkg/apis/bmc/v1beta1"
	"github.com/spidernet-io/bmc/pkg/constants"
)

// Manager handles deployment operations
type Manager struct {
	client client.Client
	scheme *runtime.Scheme
}

// NewManager creates a new deployment manager
func NewManager(client client.Client, scheme *runtime.Scheme) *Manager {
	return &Manager{
		client: client,
		scheme: scheme,
	}
}

// CreateOrUpdate creates or updates the agent deployment
func (m *Manager) CreateOrUpdate(ctx context.Context, agent *bmcv1beta1.ClusterAgent, namespace, defaultImage string) error {
	deployment := m.buildDeployment(agent, namespace, defaultImage)

	// Set owner reference
	if err := controllerutil.SetControllerReference(agent, deployment, m.scheme); err != nil {
		return err
	}

	// Create or update deployment
	err := m.client.Get(ctx, client.ObjectKey{Name: deployment.Name, Namespace: namespace}, &appsv1.Deployment{})
	if err != nil {
		if errors.IsNotFound(err) {
			return m.client.Create(ctx, deployment)
		}
		return err
	}
	return m.client.Update(ctx, deployment)
}

// buildDeployment builds the agent deployment
func (m *Manager) buildDeployment(agent *bmcv1beta1.ClusterAgent, namespace, defaultImage string) *appsv1.Deployment {
	labels := map[string]string{
		constants.LabelApp:         constants.LabelValueBMCAgent,
		constants.LabelController:  agent.Name,
		constants.LabelClusterName: agent.Spec.ClusterName,
	}

	replicas := agent.Spec.Replicas
	if replicas == 0 {
		replicas = 1
	}

	image := defaultImage
	if agent.Spec.Image != "" {
		image = agent.Spec.Image
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.AgentNamePrefix + agent.Spec.ClusterName,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
					Annotations: map[string]string{
						constants.NetworkAnnotationKey: agent.Spec.Interface,
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: constants.AgentNamePrefix + agent.Spec.ClusterName,
					Containers: []corev1.Container{{
						Name:  "bmc-agent",
						Image: image,
						Env: []corev1.EnvVar{
							{
								Name:  constants.EnvClusterName,
								Value: agent.Spec.ClusterName,
							},
						},
						Ports: []corev1.ContainerPort{{
							ContainerPort: constants.PortNumber,
							Name:         constants.PortHealth,
							Protocol:     corev1.Protocol(constants.PortProtocol),
						}},
						LivenessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								HTTPGet: &corev1.HTTPGetAction{
									Path: constants.HealthCheckPath,
									Port: intstr.FromString(constants.PortHealth),
								},
							},
							InitialDelaySeconds: 15,
							PeriodSeconds:      20,
							TimeoutSeconds:     5,
							FailureThreshold:   3,
						},
						ReadinessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								HTTPGet: &corev1.HTTPGetAction{
									Path: constants.HealthCheckPath,
									Port: intstr.FromString(constants.PortHealth),
								},
							},
							InitialDelaySeconds: 5,
							PeriodSeconds:      10,
							TimeoutSeconds:     5,
							FailureThreshold:   3,
						},
					}},
				},
			},
		},
	}
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
