package rbac

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	bmcv1beta1 "github.com/spidernet-io/bmc/pkg/apis/bmc/v1beta1"
	"github.com/spidernet-io/bmc/pkg/controller/template"
)

// Manager handles RBAC operations
type Manager struct {
	client      client.Client
	scheme      *runtime.Scheme
	tmplManager *template.Manager
}

// NewManager creates a new RBAC manager
func NewManager(client client.Client, scheme *runtime.Scheme) *Manager {
	return &Manager{
		client:      client,
		scheme:      scheme,
		tmplManager: template.NewManager("/etc/bmc/templates"),
	}
}

// ReconcileRBAC reconciles all RBAC resources for a ClusterAgent
func (m *Manager) ReconcileRBAC(ctx context.Context, clusterAgent *bmcv1beta1.ClusterAgent) error {
	// Prepare template data
	data := map[string]interface{}{
		"Name":              fmt.Sprintf("%s-agent", clusterAgent.Name),
		"Namespace":         clusterAgent.Namespace,
		"ClusterName":       clusterAgent.Spec.ClusterName,
		"ServiceAccountName": fmt.Sprintf("%s-agent-sa", clusterAgent.Name),
		"RoleName":          fmt.Sprintf("%s-agent-role", clusterAgent.Name),
	}

	// Reconcile ServiceAccount
	if err := m.reconcileResource(ctx, clusterAgent, "agent-serviceaccount.yaml", &corev1.ServiceAccount{}, data); err != nil {
		return fmt.Errorf("failed to reconcile ServiceAccount: %v", err)
	}

	// Reconcile Role
	if err := m.reconcileResource(ctx, clusterAgent, "agent-role.yaml", &rbacv1.Role{}, data); err != nil {
		return fmt.Errorf("failed to reconcile Role: %v", err)
	}

	// Reconcile RoleBinding
	if err := m.reconcileResource(ctx, clusterAgent, "agent-rolebinding.yaml", &rbacv1.RoleBinding{}, data); err != nil {
		return fmt.Errorf("failed to reconcile RoleBinding: %v", err)
	}

	return nil
}

// reconcileResource reconciles a single resource from a template
func (m *Manager) reconcileResource(ctx context.Context, clusterAgent *bmcv1beta1.ClusterAgent, templateName string, obj runtime.Object, data map[string]interface{}) error {
	// Render resource from template
	unstructured, err := m.tmplManager.RenderYAML(templateName, data)
	if err != nil {
		return fmt.Errorf("failed to render template %s: %v", templateName, err)
	}

	// Convert to concrete type
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured.Object, obj); err != nil {
		return fmt.Errorf("failed to convert unstructured to object: %v", err)
	}

	// Set controller reference
	if err := controllerutil.SetControllerReference(clusterAgent, obj.(client.Object), m.scheme); err != nil {
		return fmt.Errorf("failed to set controller reference: %v", err)
	}

	// Get object key
	key := client.ObjectKey{
		Name:      unstructured.GetName(),
		Namespace: unstructured.GetNamespace(),
	}

	// Create or update the resource
	existing := obj.DeepCopyObject().(client.Object)
	err = m.client.Get(ctx, key, existing)
	if err != nil {
		if errors.IsNotFound(err) {
			// Create new resource
			if err := m.client.Create(ctx, obj.(client.Object)); err != nil {
				return fmt.Errorf("failed to create resource: %v", err)
			}
		} else {
			return fmt.Errorf("failed to get resource: %v", err)
		}
	} else {
		// Update existing resource
		if err := m.client.Update(ctx, obj.(client.Object)); err != nil {
			return fmt.Errorf("failed to update resource: %v", err)
		}
	}

	return nil
}
