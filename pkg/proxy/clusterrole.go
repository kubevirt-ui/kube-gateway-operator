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

package proxy

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ocgatev1beta1 "github.com/yaacov/oc-gate-operator/api/v1beta1"
)

// ClusterRole is a
func ClusterRole(s *ocgatev1beta1.GateServer) (*rbacv1.ClusterRole, error) {
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
				Verbs:           s.Spec.ServiceAccountVerbs,
				APIGroups:       s.Spec.ServiceAccountAPIGroups,
				Resources:       s.Spec.ServiceAccountResources,
				ResourceNames:   s.Spec.ServiceAccountResourceNames,
				NonResourceURLs: s.Spec.ServiceAccountNonResourceURLs,
			},
		},
	}
	//controllerutil.SetControllerReference(s, role, r.Scheme)

	return role, nil
}
