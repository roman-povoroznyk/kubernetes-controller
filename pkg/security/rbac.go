// pkg/security/rbac.go
package security

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// RBACManager manages RBAC resources for the controller
type RBACManager struct {
	client    kubernetes.Interface
	namespace string
}

// NewRBACManager creates a new RBAC manager
func NewRBACManager(client kubernetes.Interface, namespace string) *RBACManager {
	return &RBACManager{
		client:    client,
		namespace: namespace,
	}
}

// EnsureServiceAccount ensures the service account exists
func (r *RBACManager) EnsureServiceAccount(ctx context.Context, name string) error {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: r.namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":     "k6s-controller",
				"app.kubernetes.io/instance": name,
			},
		},
	}

	_, err := r.client.CoreV1().ServiceAccounts(r.namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		_, err = r.client.CoreV1().ServiceAccounts(r.namespace).Create(ctx, sa, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create service account: %w", err)
		}
	}

	return nil
}

// EnsureClusterRole ensures the cluster role exists with minimal permissions
func (r *RBACManager) EnsureClusterRole(ctx context.Context, name string) error {
	cr := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"app.kubernetes.io/name": "k6s-controller",
			},
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"apps"},
				Resources: []string{"deployments"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"events"},
				Verbs:     []string{"create", "patch"},
			},
			{
				APIGroups: []string{"coordination.k8s.io"},
				Resources: []string{"leases"},
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
			},
		},
	}

	_, err := r.client.RbacV1().ClusterRoles().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		_, err = r.client.RbacV1().ClusterRoles().Create(ctx, cr, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create cluster role: %w", err)
		}
	}

	return nil
}

// EnsureClusterRoleBinding ensures the cluster role binding exists
func (r *RBACManager) EnsureClusterRoleBinding(ctx context.Context, name, serviceAccountName, clusterRoleName string) error {
	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"app.kubernetes.io/name": "k6s-controller",
			},
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      serviceAccountName,
				Namespace: r.namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     clusterRoleName,
		},
	}

	_, err := r.client.RbacV1().ClusterRoleBindings().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		_, err = r.client.RbacV1().ClusterRoleBindings().Create(ctx, crb, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create cluster role binding: %w", err)
		}
	}

	return nil
}

// CleanupRBAC removes all RBAC resources
func (r *RBACManager) CleanupRBAC(ctx context.Context, name string) error {
	// Delete ClusterRoleBinding
	err := r.client.RbacV1().ClusterRoleBindings().Delete(ctx, name+"-binding", metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete cluster role binding: %w", err)
	}

	// Delete ClusterRole
	err = r.client.RbacV1().ClusterRoles().Delete(ctx, name+"-role", metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete cluster role: %w", err)
	}

	// Delete ServiceAccount
	err = r.client.CoreV1().ServiceAccounts(r.namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete service account: %w", err)
	}

	return nil
}
