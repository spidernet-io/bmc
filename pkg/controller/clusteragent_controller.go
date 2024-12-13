package controller

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	bmcv1beta1 "github.com/spidernet-io/bmc/pkg/apis/bmc/v1beta1"
)

// ClusterAgentReconciler reconciles a ClusterAgent object
type ClusterAgentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterAgentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&bmcv1beta1.ClusterAgent{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ClusterAgentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling ClusterAgent")

	// Fetch the ClusterAgent instance
	clusterAgent := &bmcv1beta1.ClusterAgent{}
	if err := r.Get(ctx, req.NamespacedName, clusterAgent); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Initialize status if needed
	if clusterAgent.Status.Ready == false {
		clusterAgent.Status.Ready = false
		clusterAgent.Status.AllocatedIPCount = 0
		clusterAgent.Status.TotalIPCount = 0
		if clusterAgent.Status.AllocatedIPs == nil {
			clusterAgent.Status.AllocatedIPs = make(map[string]string)
		}
		if err := r.Status().Update(ctx, clusterAgent); err != nil {
			logger.Error(err, "Failed to update ClusterAgent status")
			return ctrl.Result{}, err
		}
	}

	// Check if deployment exists, if not create it
	deployment := &appsv1.Deployment{}
	err := r.Get(ctx, req.NamespacedName, deployment)
	if err != nil && errors.IsNotFound(err) {
		// Define and create a new deployment
		dep := r.deploymentForClusterAgent(clusterAgent)
		if err = r.Create(ctx, dep); err != nil {
			logger.Error(err, "Failed to create Deployment")
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		logger.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *ClusterAgentReconciler) deploymentForClusterAgent(agent *bmcv1beta1.ClusterAgent) *appsv1.Deployment {
	labels := map[string]string{
		"app":        "bmc-agent",
		"controller": agent.Name,
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      agent.Name,
			Namespace: agent.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "bmc-agent",
						Image: "bmc-agent:latest", // You should replace this with your actual image
						Env: []corev1.EnvVar{
							{
								Name:  "CLUSTER_NAME",
								Value: agent.Spec.ClusterName,
							},
							{
								Name:  "INTERFACE",
								Value: agent.Spec.Interface,
							},
						},
					}},
				},
			},
		},
	}

	// Set the owner reference
	ctrl.SetControllerReference(agent, dep, r.Scheme)
	return dep
}
