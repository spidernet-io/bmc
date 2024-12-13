package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	bmcv1beta1 "github.com/spidernet-io/bmc/pkg/apis/bmc/v1beta1"
)

// ClusterAgentReconciler reconciles a ClusterAgent object
type ClusterAgentReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (r *ClusterAgentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling ClusterAgent", "namespace", req.Namespace, "name", req.Name)

	// Fetch the ClusterAgent instance
	clusterAgent := &bmcv1beta1.ClusterAgent{}
	err := r.Get(ctx, req.NamespacedName, clusterAgent)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("ClusterAgent resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get ClusterAgent")
		return ctrl.Result{}, err
	}

	// Check if the deployment already exists, if not create a new one
	found := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: clusterAgent.Name, Namespace: clusterAgent.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := r.deploymentForClusterAgent(clusterAgent)
		logger.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.Create(ctx, dep)
		if err != nil {
			logger.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		logger.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	// Update the ClusterAgent status
	if found.Status.ReadyReplicas == found.Status.Replicas {
		if !clusterAgent.Status.Ready {
			clusterAgent.Status.Ready = true
			err = r.Status().Update(ctx, clusterAgent)
			if err != nil {
				logger.Error(err, "Failed to update ClusterAgent status")
				return ctrl.Result{}, err
			}
		}
	} else {
		if clusterAgent.Status.Ready {
			clusterAgent.Status.Ready = false
			err = r.Status().Update(ctx, clusterAgent)
			if err != nil {
				logger.Error(err, "Failed to update ClusterAgent status")
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{RequeueAfter: time.Second * 10}, nil
}

// deploymentForClusterAgent returns a ClusterAgent Deployment object
func (r *ClusterAgentReconciler) deploymentForClusterAgent(m *bmcv1beta1.ClusterAgent) *appsv1.Deployment {
	ls := labelsForClusterAgent(m.Name)
	replicas := int32(1)

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
			Labels:    ls,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
					Annotations: map[string]string{
						"k8s.v1.cni.cncf.io/networks": m.Spec.Interface,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: "spidernet-io/bmc/server:latest",
						Name:  "agent",
						Ports: []corev1.ContainerPort{{
							ContainerPort: 8000,
							Name:         "http",
							Protocol:     corev1.ProtocolTCP,
						}},
						LivenessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								HTTPGet: &corev1.HTTPGetAction{
									Path: "/healthz",
									Port: intstr.FromInt(8000),
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
									Path: "/readyz",
									Port: intstr.FromInt(8000),
								},
							},
							InitialDelaySeconds: 5,
							PeriodSeconds:      10,
							TimeoutSeconds:     5,
							FailureThreshold:   3,
						},
						Env: []corev1.EnvVar{
							{
								Name:  "ClusterName",
								Value: m.Spec.ClusterName,
							},
						},
					}},
				},
			},
		},
	}

	// Set ClusterAgent instance as the owner and controller
	controllerutil.SetControllerReference(m, dep, r.Scheme)
	return dep
}

func labelsForClusterAgent(name string) map[string]string {
	return map[string]string{
		"app":                        "cluster-agent",
		"app.kubernetes.io/name":     "cluster-agent",
		"app.kubernetes.io/instance": name,
		"app.kubernetes.io/part-of":  "bmc",
		"controller":                 name,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterAgentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&bmcv1beta1.ClusterAgent{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
