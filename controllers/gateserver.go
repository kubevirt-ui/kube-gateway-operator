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
	"context"
	"fmt"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ocgatev1beta1 "github.com/yaacov/oc-gate-operator/api/v1beta1"
)

func setServerCondition(s *ocgatev1beta1.GateServer, reason string, err error) {
	t := metav1.Time{Time: time.Now()}
	s.Status.Phase = "Error"
	condition := metav1.Condition{
		Type:               "Error",
		Status:             "True",
		Reason:             reason,
		Message:            fmt.Sprintf("%s", err),
		LastTransitionTime: t,
	}
	s.Status.Conditions = []metav1.Condition{condition}
}

func (r *GateServerReconciler) service(s *ocgatev1beta1.GateServer) (*corev1.Service, error) {
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
		},
	}

	controllerutil.SetControllerReference(s, service, r.Scheme)

	return service, nil
}

func (r *GateServerReconciler) route(s *ocgatev1beta1.GateServer) (*routev1.Route, error) {
	labels := map[string]string{
		"app": s.Name,
	}

	route := &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.Name,
			Namespace: s.Namespace,
			Labels:    labels,
		},
		Spec: routev1.RouteSpec{
			Host: s.Spec.Route,
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: s.Name,
			},
			TLS: &routev1.TLSConfig{
				Termination: routev1.TLSTerminationReencrypt,
			},
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromInt(8080),
			},
			WildcardPolicy: routev1.WildcardPolicyNone,
		},
	}

	controllerutil.SetControllerReference(s, route, r.Scheme)

	return route, nil
}

func (r *GateServerReconciler) serviceaccount(s *ocgatev1beta1.GateServer) (*corev1.ServiceAccount, error) {
	labels := map[string]string{
		"app": s.Name,
	}

	serviceaccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.Name,
			Namespace: s.Namespace,
			Labels:    labels,
		},
		Secrets: []corev1.ObjectReference{
			{
				Name: fmt.Sprintf("%s-jwt-secret", s.Name),
			},
		},
	}
	controllerutil.SetControllerReference(s, serviceaccount, r.Scheme)

	return serviceaccount, nil
}

func (r *GateServerReconciler) role(s *ocgatev1beta1.GateServer) (*rbacv1.Role, error) {
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

	controllerutil.SetControllerReference(s, role, r.Scheme)

	return role, nil
}

func (r *GateServerReconciler) rolebinding(s *ocgatev1beta1.GateServer) (*rbacv1.RoleBinding, error) {
	labels := map[string]string{
		"app": s.Name,
	}

	rolebinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.Name,
			Namespace: s.Namespace,
			Labels:    labels,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind: "ServiceAccount",
				Name: s.Name,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     s.Name,
		},
	}

	controllerutil.SetControllerReference(s, rolebinding, r.Scheme)

	return rolebinding, nil
}

func (r *GateServerReconciler) deployment(s *ocgatev1beta1.GateServer) (*appsv1.Deployment, error) {
	replicas := int32(1)
	labels := map[string]string{
		"app": s.Name,
	}
	matchlabels := map[string]string{
		"app": s.Name,
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.Name,
			Namespace: s.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: matchlabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: matchlabels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: s.Spec.Image,
						Name:  "oc-gate",

						Ports: []corev1.ContainerPort{{
							ContainerPort: 8080,
							Name:          "oc-gate-https",
						}},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "oc-gate-secret",
								MountPath: "/secrets",
							},
							{
								Name:      "oc-gate-jwt-secret",
								MountPath: "/jwt-secret",
							},
						},
						Command: []string{
							"./oc-gate",
							fmt.Sprintf("-api-server=%s", s.Spec.APIURL),
							"-ca-file=/var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt",
							"-cert-file=/secrets/tls.crt",
							"-key-file=/secrets/tls.key",
							fmt.Sprintf("-base-address=https://%s", s.Spec.Route),
							"-listen=https://0.0.0.0:8080",
							"-jwt-token-key-file=/jwt-secret/cert.pem",
							"-k8s-bearer-token-file=/var/run/secrets/kubernetes.io/serviceaccount/token",
							fmt.Sprintf("-k8s-bearer-token-passthrough=%v", s.Spec.PassThrough),
						},
					}},

					Volumes: []corev1.Volume{
						{
							Name: "oc-gate-secret",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: fmt.Sprintf("%s-secret", s.Name),
								},
							},
						},
						{
							Name: "oc-gate-jwt-secret",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "oc-gate-jwt-secret",
								},
							},
						},
					},

					ServiceAccountName: s.Name,
				},
			},
		},
	}

	controllerutil.SetControllerReference(s, deployment, r.Scheme)

	return deployment, nil
}

func (r *GateServerReconciler) buildServer(ctx context.Context, s *ocgatev1beta1.GateServer) error {
	// Create the service and route
	se, _ := r.service(s)
	err := r.Client.Create(ctx, se)
	if err != nil {
		return err
	}

	route, _ := r.route(s)
	err = r.Client.Create(ctx, route)
	if err != nil {
		return err
	}

	// Create the service account and roles
	sa, _ := r.serviceaccount(s)
	err = r.Client.Create(ctx, sa)
	if err != nil {
		return err
	}
	if s.Spec.AdminNamespaced {
		role, _ := r.role(s)
		err = r.Client.Create(ctx, role)
		if err != nil {
			return err
		}
	} else {
		role, _ := r.clusterrole(s)
		err = r.Client.Create(ctx, role)
		if err != nil {
			return err
		}
	}
	rolebinding, _ := r.rolebinding(s)
	err = r.Client.Create(ctx, rolebinding)
	if err != nil {
		return err
	}

	// Create the gate service
	dep, _ := r.deployment(s)
	err = r.Client.Create(ctx, dep)
	if err != nil {
		return err
	}

	return nil
}
