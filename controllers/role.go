package controllers

import (
	"strings"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	kubegatewayv1beta1 "github.com/kubevirt-ui/kube-gateway-operator/api/v1beta1"
)

// Role creates a role resource
func (r *GateServerReconciler) Role(s *kubegatewayv1beta1.GateServer) (*rbacv1.Role, error) {
	var verbs []string
	var resources []string

	labels := map[string]string{
		"app": s.Name,
	}

	if s.Spec.AdminRole == "admin" {
		verbs = []string{"get", "list", "watch", "create", "delete", "patch", "update"}
	} else {
		verbs = []string{"get", "list", "watch"}
	}
	if s.Spec.AdminResources == "" {
		resources = []string{"*"}
	} else {
		resources = strings.Split(s.Spec.AdminResources, ",")
	}

	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.Name,
			Namespace: s.Namespace,
			Labels:    labels,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"*"},
				Resources: resources,
				Verbs:     verbs,
			},
		},
	}

	controllerutil.SetControllerReference(s, role, r.Scheme)

	return role, nil
}
