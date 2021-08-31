package controllers

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	kubegatewayv1beta1 "github.com/kubevirt-ui/kube-gateway-operator/api/v1beta1"
)

// CreateResources creates resources needed to run the gateway proxy
// - secrets
// - service
// - service account
// - role
// - rolebinding
// - route (FIXME: requirs openshift)
func (r *GateServerReconciler) CreateResources(ctx context.Context, gateserver *kubegatewayv1beta1.GateServer) (ctrl.Result, error) {
	var err error

	// Take time
	t := metav1.Time{Time: time.Now()}

	// Create the JWT secret
	r.Log.Info("Create JWT secret.")
	secret, _ := r.Secret(gateserver)
	if err := r.Client.Create(ctx, secret); err != nil {
		r.Log.Info("Failed to create service.", "err", err)

		gateserver.Status.Phase = "Error"
		condition := metav1.Condition{
			Type:               "SecretCreated",
			Status:             "False",
			Reason:             "FailedCreateSecret",
			Message:            fmt.Sprintf("%s", err),
			LastTransitionTime: t,
		}
		gateserver.Status.Conditions = append(gateserver.Status.Conditions, condition)
		if err := r.Status().Update(ctx, gateserver); err != nil {
			r.Log.Info("Failed to update status", "err", err)
		}

		return ctrl.Result{}, nil
	}

	// Create the service and route
	se, _ := r.Service(gateserver)
	err = r.Client.Create(ctx, se)
	if err != nil {
		r.Log.Info("Failed to create service.", "err", err)

		gateserver.Status.Phase = "Error"
		condition := metav1.Condition{
			Type:               "ServiceCreated",
			Status:             "False",
			Reason:             "FailedCreateService",
			Message:            fmt.Sprintf("%s", err),
			LastTransitionTime: t,
		}
		gateserver.Status.Conditions = append(gateserver.Status.Conditions, condition)
		if err := r.Status().Update(ctx, gateserver); err != nil {
			r.Log.Info("Failed to update status", "err", err)
		}

		return ctrl.Result{}, err
	}

	// Create the service account and roles
	sa, _ := r.ServiceAccount(gateserver)
	err = r.Client.Create(ctx, sa)
	if err != nil {
		r.Log.Info("Failed to create serviceaccount.", "err", err)

		gateserver.Status.Phase = "Error"
		condition := metav1.Condition{
			Type:               "ServiceaccountCreated",
			Status:             "False",
			Reason:             "FailedCreateServiceaccount",
			Message:            fmt.Sprintf("%s", err),
			LastTransitionTime: t,
		}
		gateserver.Status.Conditions = append(gateserver.Status.Conditions, condition)
		if err := r.Status().Update(ctx, gateserver); err != nil {
			r.Log.Info("Failed to update status", "err", err)
		}

		return ctrl.Result{}, err
	}

	role, _ := r.Role(gateserver)
	err = r.Client.Create(ctx, role)
	if err != nil {
		r.Log.Info("Failed to create role.", "err", err)

		gateserver.Status.Phase = "Error"
		condition := metav1.Condition{
			Type:               "RoleCreated",
			Status:             "False",
			Reason:             "FailedCreateRole",
			Message:            fmt.Sprintf("%s", err),
			LastTransitionTime: t,
		}
		gateserver.Status.Conditions = append(gateserver.Status.Conditions, condition)
		if err := r.Status().Update(ctx, gateserver); err != nil {
			r.Log.Info("Failed to update status", "err", err)
		}

		return ctrl.Result{}, err
	}

	rolebinding, _ := r.RoleBinding(gateserver)
	err = r.Client.Create(ctx, rolebinding)
	if err != nil {
		r.Log.Info("Failed to create rolebinding.", "err", err)

		gateserver.Status.Phase = "Error"
		condition := metav1.Condition{
			Type:               "RolebindingCreated",
			Status:             "False",
			Reason:             "FailedCreateRolebinding",
			Message:            fmt.Sprintf("%s", err),
			LastTransitionTime: t,
		}
		gateserver.Status.Conditions = append(gateserver.Status.Conditions, condition)
		if err := r.Status().Update(ctx, gateserver); err != nil {
			r.Log.Info("Failed to update status", "err", err)
		}

		return ctrl.Result{}, err
	}

	route, _ := r.Route(gateserver)
	err = r.Client.Create(ctx, route)
	if err != nil {
		r.Log.Info("Failed to create route.", "err", err)

		gateserver.Status.Phase = "Error"
		condition := metav1.Condition{
			Type:               "RouteCreated",
			Status:             "False",
			Reason:             "FailedCreateRoute",
			Message:            fmt.Sprintf("%s", err),
			LastTransitionTime: t,
		}
		gateserver.Status.Conditions = append(gateserver.Status.Conditions, condition)
		if err := r.Status().Update(ctx, gateserver); err != nil {
			r.Log.Info("Failed to update status", "err", err)
		}

		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}
