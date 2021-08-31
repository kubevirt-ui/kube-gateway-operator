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
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	kubegatewayv1beta1 "github.com/kubevirt-ui/kube-gateway-operator/api/v1beta1"
)

// Service creates a service resource
// This service shoult be load balanced using a node port open to outside the cluster
func (r *GateServerReconciler) Service(s *kubegatewayv1beta1.GateServer) (*corev1.Service, error) {
	labels := map[string]string{
		"app": s.Name,
	}
	annotations := map[string]string{
		"service.alpha.openshift.io/serving-cert-secret-name": fmt.Sprintf("%s-secret", s.Name),
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        s.Name,
			Namespace:   s.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Port:       8080,
					Protocol:   corev1.ProtocolTCP,
					TargetPort: intstr.FromInt(8080),
				},
			},
			Type: corev1.ServiceTypeLoadBalancer,
		},
	}

	controllerutil.SetControllerReference(s, service, r.Scheme)

	return service, nil
}
