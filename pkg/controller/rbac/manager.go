package rbac

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	bmcv1beta1 "github.com/spidernet-io/bmc/pkg/apis/bmc/v1beta1"
)

// Manager manages RBAC resources for ClusterAgent
type Manager struct {
	client client.Client
	scheme *runtime.Scheme
}

// NewManager creates a new RBAC manager
func NewManager(client client.Client, scheme *runtime.Scheme) *Manager {
	return &Manager{
		client: client,
		scheme: scheme,
	}
}

// formatName formats a name to be RFC 1123 compliant
func formatName(name string) string {
	// Replace any uppercase letters with lowercase
	name = strings.ToLower(name)
	// Replace any non-alphanumeric characters with '-'
	name = regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(name, "-")
	// Ensure the name starts and ends with an alphanumeric character
	name = strings.Trim(name, "-")
	// If the name is empty after trimming, use a default name
	if name == "" {
		name = "agent"
	}
	return name
}

// ReconcileServiceAccount reconciles the ServiceAccount for the ClusterAgent
func (m *Manager) ReconcileServiceAccount(ctx context.Context, agent *bmcv1beta1.ClusterAgent, namespace string) error {
	saName := formatName(fmt.Sprintf("agent-%s", agent.Spec.ClusterName))

	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      saName,
			Namespace: namespace,
		},
	}

	if err := controllerutil.SetControllerReference(agent, sa, m.scheme); err != nil {
		return fmt.Errorf("failed to set controller reference: %w", err)
	}

	_, err := controllerutil.CreateOrUpdate(ctx, m.client, sa, func() error {
		return nil // No updates needed for ServiceAccount
	})

	return err
}

// ReconcileRole reconciles the Role for the ClusterAgent
func (m *Manager) ReconcileRole(ctx context.Context, agent *bmcv1beta1.ClusterAgent, namespace string) error {
	roleName := formatName(fmt.Sprintf("agent-%s", agent.Spec.ClusterName))
	
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      roleName,
			Namespace: namespace,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"pods", "services", "endpoints", "persistentvolumeclaims", "events", "configmaps", "secrets"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{"apps"},
				Resources: []string{"deployments", "daemonsets", "replicasets", "statefulsets"},
				Verbs:     []string{"*"},
			},
		},
	}

	if err := controllerutil.SetControllerReference(agent, role, m.scheme); err != nil {
		return fmt.Errorf("failed to set controller reference: %w", err)
	}

	_, err := controllerutil.CreateOrUpdate(ctx, m.client, role, func() error {
		role.Rules = []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"pods", "services", "endpoints", "persistentvolumeclaims", "events", "configmaps", "secrets"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups: []string{"apps"},
				Resources: []string{"deployments", "daemonsets", "replicasets", "statefulsets"},
				Verbs:     []string{"*"},
			},
		}
		return nil
	})

	return err
}

// ReconcileRoleBinding reconciles the RoleBinding for the ClusterAgent
func (m *Manager) ReconcileRoleBinding(ctx context.Context, agent *bmcv1beta1.ClusterAgent, namespace string) error {
	roleName := formatName(fmt.Sprintf("agent-%s", agent.Spec.ClusterName))
	saName := formatName(fmt.Sprintf("agent-%s", agent.Spec.ClusterName))

	binding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      roleName,
			Namespace: namespace,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      saName,
				Namespace: namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     roleName,
		},
	}

	if err := controllerutil.SetControllerReference(agent, binding, m.scheme); err != nil {
		return fmt.Errorf("failed to set controller reference: %w", err)
	}

	_, err := controllerutil.CreateOrUpdate(ctx, m.client, binding, func() error {
		binding.Subjects = []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      saName,
				Namespace: namespace,
			},
		}
		binding.RoleRef = rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     roleName,
		}
		return nil
	})

	return err
}

// CleanupRBACResources cleans up all RBAC resources for the ClusterAgent
func (m *Manager) CleanupRBACResources(ctx context.Context, agent *bmcv1beta1.ClusterAgent, namespace string) error {
	roleName := formatName(fmt.Sprintf("agent-%s", agent.Spec.ClusterName))
	saName := formatName(fmt.Sprintf("agent-%s", agent.Spec.ClusterName))

	// Delete RoleBinding
	binding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      roleName,
			Namespace: namespace,
		},
	}
	if err := m.client.Delete(ctx, binding); err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete RoleBinding: %w", err)
	}

	// Delete Role
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      roleName,
			Namespace: namespace,
		},
	}
	if err := m.client.Delete(ctx, role); err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete Role: %w", err)
	}

	// Delete ServiceAccount
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      saName,
			Namespace: namespace,
		},
	}
	if err := m.client.Delete(ctx, sa); err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete ServiceAccount: %w", err)
	}

	return nil
}
