package controllers

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	kubegatewayv1beta1 "github.com/kubevirt-ui/kube-gateway-operator/api/v1beta1"
)

// Deployment creates a deployment resource
func (r *GateServerReconciler) Deployment(s *kubegatewayv1beta1.GateServer) (*appsv1.Deployment, error) {
	image := s.Spec.IMG
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
						Image: image,
						Name:  "kube-gateway",

						Ports: []corev1.ContainerPort{{
							ContainerPort: 8080,
							Name:          "https",
						}},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "serving-cert",
								MountPath: "/var/run/secrets/serving-cert",
							},
						},
						Command: []string{
							"./kube-gateway",
							"-api-server=https://kubernetes.default.svc",
							"-gateway-listen=https://0.0.0.0:8080",
							"-api-server-ca-file=/var/run/secrets/kubernetes.io/serviceaccount/ca.crt",
							"-api-server-bearer-token-file=/var/run/secrets/kubernetes.io/serviceaccount/token",
							"-gateway-key-file=/var/run/secrets/serving-cert/tls.key",
							"-gateway-cert-file=/var/run/secrets/serving-cert/tls.crt",
							fmt.Sprintf("-jwt-public-key-name=%s-jwt-secret", s.Name),
							fmt.Sprintf("-jwt-public-key-namespace=%s", s.Namespace),
							"-jwt-request-enable=true",
							fmt.Sprintf("-jwt-private-key-name=%s-jwt-secret", s.Name),
							fmt.Sprintf("-jwt-private-key-namespace=%s", s.Namespace),
						},
					}},

					Volumes: []corev1.Volume{
						{
							Name: "serving-cert",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: fmt.Sprintf("%s-secret", s.Name),
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
