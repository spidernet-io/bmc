package controller

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	bmcv1beta1 "github.com/spidernet-io/bmc/pkg/apis/bmc/v1beta1"
)

// BMCServerReconciler reconciles a BMCServer object
type BMCServerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=bmc.io,resources=bmcservers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=bmc.io,resources=bmcservers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

func (r *BMCServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling BMCServer", "namespacedName", req.NamespacedName)

	// Fetch the BMCServer instance
	bmcServer := &bmcv1beta1.BMCServer{}
	err := r.Get(ctx, req.NamespacedName, bmcServer)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("BMCServer resource not found, ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get BMCServer")
		return ctrl.Result{}, err
	}

	// Check if the deployment already exists, if not create a new one
	deployment := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: bmcServer.Name, Namespace: bmcServer.Namespace}, deployment)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := r.deploymentForBMCServer(bmcServer)
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

	// Update BMCServer status based on deployment status
	ready := isDeploymentReady(deployment)
	if ready != bmcServer.Status.Ready {
		bmcServer.Status.Ready = ready
		err = r.Status().Update(ctx, bmcServer)
		if err != nil {
			logger.Error(err, "Failed to update BMCServer status")
			return ctrl.Result{}, err
		}
		logger.Info("Updated BMCServer status", "ready", ready)
	}

	return ctrl.Result{}, nil
}

// deploymentForBMCServer returns a BMCServer Deployment object
func (r *BMCServerReconciler) deploymentForBMCServer(m *bmcv1beta1.BMCServer) *appsv1.Deployment {
	ls := labelsForBMCServer(m.Name)
	replicas := int32(1)

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
			Annotations: map[string]string{
				"k8s.v1.cni.cncf.io/networks": m.Spec.Interface,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: "bmc/server:latest",
						Name:  "bmcserver",
						Env: []corev1.EnvVar{{
							Name:  "ClusterName",
							Value: m.Spec.ClusterName,
						}},
					}},
				},
			},
		},
	}

	// Set BMCServer instance as the owner and controller
	ctrl.SetControllerReference(m, dep, r.Scheme)
	return dep
}

// labelsForBMCServer returns the labels for selecting the resources
func labelsForBMCServer(name string) map[string]string {
	return map[string]string{
		"app":        "bmcserver",
		"instance":   name,
		"created-by": "bmc-operator",
	}
}

// isDeploymentReady checks if all pods in the deployment are running
func isDeploymentReady(deployment *appsv1.Deployment) bool {
	return deployment.Status.ReadyReplicas == *deployment.Spec.Replicas
}

// SetupWithManager sets up the controller with the Manager.
func (r *BMCServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&bmcv1beta1.BMCServer{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
