/*
Copyright 2021 Yaacov Zamir.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"strings"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ocgatev1beta1 "github.com/yaacov/oc-gate-operator/api/v1beta1"
)

func (r *GateServerReconciler) clusterrole(s *ocgatev1beta1.GateServer) (*rbacv1.ClusterRole, error) {
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

	role := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   s.Name,
			Labels: labels,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"*"},
				Resources: resources,
				Verbs:     verbs,
			},
		},
	}

	return role, nil
}

func (r *GateServerReconciler) clusterrolebinding(s *ocgatev1beta1.GateServer) (*rbacv1.ClusterRoleBinding, error) {
	labels := map[string]string{
		"app": s.Name,
	}

	rolebinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   s.Name,
			Labels: labels,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Namespace: s.Namespace,
				Name:      s.Name,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     s.Name,
		},
	}

	return rolebinding, nil
}
