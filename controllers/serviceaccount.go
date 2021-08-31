package controllers

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	kubegatewayv1beta1 "github.com/kubevirt-ui/kube-gateway-operator/api/v1beta1"
)

// ServiceAccount creates a service account resource
// the secret part will create a secret holding a token and cridentials to log
// into the API server using this service account
func (r *GateServerReconciler) ServiceAccount(s *kubegatewayv1beta1.GateServer) (*corev1.ServiceAccount, error) {
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
				Name: fmt.Sprintf("%s-secret", s.Name),
			},
		},
	}
	controllerutil.SetControllerReference(s, serviceaccount, r.Scheme)

	return serviceaccount, nil
}
