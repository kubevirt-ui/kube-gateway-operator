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
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ocgatev1beta1 "github.com/yaacov/oc-gate-operator/api/v1beta1"
)

func (r *GateTokenReconciler) serviceaccount(s *ocgatev1beta1.GateToken) (*corev1.ServiceAccount, error) {
	labels := map[string]string{
		"app": s.Name,
	}

	serviceaccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{Name: s.Name, Namespace: s.Namespace, Labels: labels},
		Secrets: []corev1.ObjectReference{
			{
				Name:      s.Name,
				Namespace: s.Namespace,
			},
		},
	}
	controllerutil.SetControllerReference(s, serviceaccount, r.Scheme)

	return serviceaccount, nil
}

func (r *GateTokenReconciler) clusterrole(s *ocgatev1beta1.GateToken) (*rbacv1.ClusterRole, error) {
	labels := map[string]string{
		"app": s.Name,
	}

	role := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   s.Name,
			Labels: labels,
		},
		Rules: []rbacv1.PolicyRule{
			{
				Verbs:           s.Spec.Verbs,
				APIGroups:       s.Spec.APIGroups,
				Resources:       s.Spec.Resources,
				ResourceNames:   s.Spec.ResourceNames,
				NonResourceURLs: s.Spec.NonResourceURLs,
			},
		},
	}
	controllerutil.SetControllerReference(s, role, r.Scheme)

	return role, nil
}

func (r *GateTokenReconciler) clusterrolebinding(s *ocgatev1beta1.GateToken) (*rbacv1.ClusterRoleBinding, error) {
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
	controllerutil.SetControllerReference(s, rolebinding, r.Scheme)

	return rolebinding, nil
}
