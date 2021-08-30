package controllers

import (
	routev1 "github.com/openshift/api/route/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	ocgatev1beta1 "github.com/yaacov/oc-gate-operator/api/v1beta1"
)

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
