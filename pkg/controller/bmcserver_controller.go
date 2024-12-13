package controller

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	bmcv1beta1 "github.com/spidernet-io/bmc/pkg/apis/bmc/v1beta1"
)

// BMCServerReconciler reconciles a BMCServer object
type BMCServerReconciler struct {
	client.Client
}

// SetupWithManager sets up the controller with the Manager.
func (r *BMCServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&bmcv1beta1.BMCServer{}).
		Complete(r)
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *BMCServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling BMCServer")

	// Fetch the BMCServer instance
	bmcServer := &bmcv1beta1.BMCServer{}
	if err := r.Get(ctx, req.NamespacedName, bmcServer); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Initialize status if needed
	if bmcServer.Status.Connected == false {
		bmcServer.Status.Connected = false
		if err := r.Status().Update(ctx, bmcServer); err != nil {
			logger.Error(err, "Failed to update BMCServer status")
			return ctrl.Result{}, err
		}
	}

	// Your reconciliation logic here
	// For example, try to connect to the BMC server and update status accordingly
	if err := r.connectToBMC(bmcServer); err != nil {
		logger.Error(err, "Failed to connect to BMC server")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *BMCServerReconciler) connectToBMC(server *bmcv1beta1.BMCServer) error {
	// Implement BMC connection logic here
	// This is just a placeholder
	return fmt.Errorf("BMC connection not implemented")
}
