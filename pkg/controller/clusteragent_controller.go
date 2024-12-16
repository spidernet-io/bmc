package controller

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	bmcv1beta1 "github.com/spidernet-io/bmc/pkg/apis/bmc/v1beta1"
	"github.com/spidernet-io/bmc/pkg/constants"
	"github.com/spidernet-io/bmc/pkg/controller/rbac"
	"github.com/spidernet-io/bmc/pkg/controller/template"
	"github.com/spidernet-io/bmc/pkg/utils"
)

const (
	finalizerName = "bmc.spidernet.io/finalizer"
)

// ClusterAgentReconciler reconciles a ClusterAgent object
type ClusterAgentReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	rbacMgr    *rbac.Manager
	deployMgr  *template.Manager
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterAgentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.rbacMgr = rbac.NewManager(r.Client, r.Scheme)
	r.deployMgr = template.NewManager(r.Client, r.Scheme)

	return ctrl.NewControllerManagedBy(mgr).
		For(&bmcv1beta1.ClusterAgent{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}

// Reconcile is part of the main kubernetes reconciliation loop
func (r *ClusterAgentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling ClusterAgent")

	// Fetch the ClusterAgent instance
	clusterAgent := &bmcv1beta1.ClusterAgent{}
	if err := r.Get(ctx, req.NamespacedName, clusterAgent); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Get controller's namespace
	controllerNS := os.Getenv(constants.EnvPodNamespace)
	if controllerNS == "" {
		logger.Error(nil, "POD_NAMESPACE environment variable not set")
		return ctrl.Result{}, fmt.Errorf("POD_NAMESPACE environment variable not set")
	}

	// Handle deletion
	if !clusterAgent.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, clusterAgent, controllerNS)
	}

	// Add finalizer if it doesn't exist
	if !utils.ContainsString(clusterAgent.Finalizers, constants.ClusterAgentFinalizer) {
		clusterAgent.Finalizers = append(clusterAgent.Finalizers, constants.ClusterAgentFinalizer)
		if err := r.Update(ctx, clusterAgent); err != nil {
			logger.Error(err, "Failed to add finalizer")
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Get agent image from environment
	agentImage := os.Getenv(constants.EnvAgentImage)
	if agentImage == "" {
		logger.Error(nil, "agentImage environment variable not set")
		return ctrl.Result{}, fmt.Errorf("agentImage environment variable not set")
	}

	// Reconcile RBAC resources
	if err := r.rbacMgr.ReconcileServiceAccount(ctx, clusterAgent, controllerNS); err != nil {
		logger.Error(err, "Failed to reconcile ServiceAccount")
		return ctrl.Result{}, err
	}

	if err := r.rbacMgr.ReconcileRole(ctx, clusterAgent, controllerNS); err != nil {
		logger.Error(err, "Failed to reconcile Role")
		return ctrl.Result{}, err
	}

	if err := r.rbacMgr.ReconcileRoleBinding(ctx, clusterAgent, controllerNS); err != nil {
		logger.Error(err, "Failed to reconcile RoleBinding")
		return ctrl.Result{}, err
	}

	// Reconcile Deployment
	deploymentName := constants.AgentNamePrefix + clusterAgent.Spec.ClusterName
	deployment := &appsv1.Deployment{}
	if err := r.Get(ctx, client.ObjectKey{Name: deploymentName, Namespace: controllerNS}, deployment); err != nil {
		if errors.IsNotFound(err) {
			if err := r.deployMgr.CreateOrUpdate(ctx, clusterAgent, controllerNS, agentImage); err != nil {
				logger.Error(err, "Failed to create Deployment")
				return ctrl.Result{}, err
			}
			logger.Info("Created new Deployment", "Name", deploymentName)
			return ctrl.Result{Requeue: true}, nil
		}
		logger.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	// Update deployment if needed
	if r.deployMgr.ShouldUpdate(deployment, clusterAgent, agentImage) {
		if err := r.deployMgr.CreateOrUpdate(ctx, clusterAgent, controllerNS, agentImage); err != nil {
			logger.Error(err, "Failed to update Deployment")
			return ctrl.Result{}, err
		}
		logger.Info("Updated Deployment", "Name", deploymentName)
	}

	// Update ClusterAgent status
	ready := r.deployMgr.IsReady(deployment)
	if clusterAgent.Status.Ready != ready {
		clusterAgent.Status.Ready = ready
		if err := r.Status().Update(ctx, clusterAgent); err != nil {
			logger.Error(err, "Failed to update ClusterAgent status")
			return ctrl.Result{}, err
		}
		logger.Info("Updated ClusterAgent status", "Ready", ready)
	}

	return ctrl.Result{}, nil
}

func (r *ClusterAgentReconciler) handleDeletion(ctx context.Context, agent *bmcv1beta1.ClusterAgent, namespace string) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Handling ClusterAgent deletion", "name", agent.Name)

	if utils.ContainsString(agent.Finalizers, constants.ClusterAgentFinalizer) {
		// Clean up resources
		deploymentName := constants.AgentNamePrefix + agent.Spec.ClusterName
		if err := r.deployMgr.Delete(ctx, deploymentName, namespace); err != nil {
			logger.Error(err, "Failed to delete Deployment")
			return ctrl.Result{}, err
		}

		if err := r.rbacMgr.CleanupRBACResources(ctx, agent, namespace); err != nil {
			logger.Error(err, "Failed to clean up RBAC resources")
			return ctrl.Result{}, err
		}

		// Remove finalizer
		agent.Finalizers = utils.RemoveString(agent.Finalizers, constants.ClusterAgentFinalizer)
		if err := r.Update(ctx, agent); err != nil {
			logger.Error(err, "Failed to remove finalizer")
			return ctrl.Result{}, err
		}
		logger.Info("Successfully removed finalizer")
	}

	return ctrl.Result{}, nil
}
