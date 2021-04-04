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
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ocgatev1beta1 "github.com/yaacov/virt-gateway-operator/api/v1beta1"
)

// Deployment is a
func Deployment(s *ocgatev1beta1.GateServer) (*appsv1.Deployment, error) {
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
					InitContainers: initContainers(s),
					Containers:     containers(s),

					Volumes: []corev1.Volume{
						{
							Name: "kube-gateway-secret",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: fmt.Sprintf("%s-secret", s.Name),
								},
							},
						},
						{
							Name: "kube-gateway-jwt-secret",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "kube-gateway-jwt-secret",
								},
							},
						},
						{
							Name: "web-application",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},

					ServiceAccountName: s.Name,
				},
			},
		},
	}

	return deployment, nil
}

func initContainers(s *ocgatev1beta1.GateServer) []corev1.Container {
	// Return nil if no web app available
	if s.Spec.WebAppImage == "" {
		return nil
	}

	containers := []corev1.Container{{
		Image: s.Spec.WebAppImage,
		Name:  "kube-gateway-web-app",

		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "web-application",
				MountPath: "/app/web",
			},
		},
		Command: []string{
			"/bin/cp", "-r", "/data/web/public", "/app/web/",
		},
	}}

	return containers
}

func containers(s *ocgatev1beta1.GateServer) []corev1.Container {
	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "kube-gateway-secret",
			MountPath: "/secrets",
		},
		{
			Name:      "kube-gateway-jwt-secret",
			MountPath: "/jwt-secret",
		},
	}

	// If using a web app, add the web app volume mount
	if s.Spec.WebAppImage != "" {
		webAppVolumeMount := corev1.VolumeMount{
			Name:      "web-application",
			MountPath: "/app/web",
		}

		volumeMounts = append(volumeMounts, webAppVolumeMount)
	}

	containers := []corev1.Container{{
		Image: s.Spec.Image,
		Name:  "kube-gateway",

		Ports: []corev1.ContainerPort{{
			ContainerPort: 8080,
			Name:          "https",
		}},
		VolumeMounts: volumeMounts,
		Command: []string{
			"./kube-gateway",
			fmt.Sprintf("-api-server=%s", s.Spec.APIURL),
			"-ca-file=/var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt",
			"-cert-file=/secrets/tls.crt",
			"-key-file=/secrets/tls.key",
			fmt.Sprintf("-base-address=https://%s", s.Spec.Route),
			"-listen=https://0.0.0.0:8080",
			"-jwt-token-key-file=/jwt-secret/cert.pem",
			"-k8s-bearer-token-file=/var/run/secrets/kubernetes.io/serviceaccount/token",
			fmt.Sprintf("-k8s-bearer-token-passthrough=%v", s.Spec.PassThrough),
			fmt.Sprintf("-oauth-client-id=%s", s.Name),
			fmt.Sprintf("-oauth-client-secret=%s-oauth-secret", s.Name),
		},
	}}

	return containers
}
