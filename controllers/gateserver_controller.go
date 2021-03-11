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

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ocgatev1beta1 "github.com/yaacov/oc-gate-operator/api/v1beta1"
)

const gateserverFinalizer = "ocgate.yaacov.com/finalizer"

// GateServerReconciler reconciles a GateServer object
type GateServerReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="route.openshift.io",resources=routes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="route.openshift.io",resources=routes/custom-host,verbs=create;patch
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=roles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="apps",resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=security.openshift.io,resources=securitycontextconstraints,resourceNames=privileged,verbs=use
// +kubebuilder:rbac:groups=ocgate.yaacov.com,resources=gateservers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ocgate.yaacov.com,resources=gateservers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=ocgate.yaacov.com,resources=gateservers/finalizers,verbs=update

// In order to grant client users access to resource, the operator it'slef need this access.
// Note: to create a new gate server with admin role, the role of this operator need to be adjusted.
// +kubebuilder:rbac:groups="*",resources="*",verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the GateServer object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *GateServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("gateserver", req.NamespacedName)

	// your logic here

	// Lookup the GateToken instance for this reconcile request
	gateserver := &ocgatev1beta1.GateServer{}
	err := r.Get(ctx, req.NamespacedName, gateserver)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			r.Log.Info("GateServer resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		r.Log.Error(err, "Failed to get GateServer.")
		return ctrl.Result{}, err
	}

	// Check if the GateServer instance is marked to be deleted, which is
	// indicated by the deletion timestamp being set.
	isGateServerMarkedToBeDeleted := gateserver.GetDeletionTimestamp() != nil
	if isGateServerMarkedToBeDeleted {
		if controllerutil.ContainsFinalizer(gateserver, gateserverFinalizer) {
			// Run finalization logic for gateserverFinalizer. If the
			// finalization logic fails, don't remove the finalizer so
			// that we can retry during the next reconciliation.
			if err := r.finalizeGateServer(gateserver); err != nil {
				return ctrl.Result{}, err
			}

			// Remove gateserverFinalizer. Once all finalizers have been
			// removed, the object will be deleted.
			controllerutil.RemoveFinalizer(gateserver, gateserverFinalizer)
			err := r.Update(ctx, gateserver)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// If token was created, exit.
	if gateserver.Status.Phase != "" {
		r.Log.Info("Old server", "id", gateserver.Name)
		return ctrl.Result{}, nil
	}

	// Take time
	t := metav1.Time{Time: time.Now()}

	// Create the service and route
	se, _ := r.service(gateserver)
	err = r.Client.Create(ctx, se)
	if err != nil {
		r.Log.Info("Failed to create service.", "err", err)

		setServerCondition(gateserver, "FailedCreateService", err)
		if err := r.Status().Update(ctx, gateserver); err != nil {
			r.Log.Info("Failed to update status", "err", err)
		}
		return ctrl.Result{}, nil
	}

	route, _ := r.route(gateserver)
	err = r.Client.Create(ctx, route)
	if err != nil {
		r.Log.Info("Failed to create route.", "err", err)

		setServerCondition(gateserver, "FailedCreateRoute", err)
		if err := r.Status().Update(ctx, gateserver); err != nil {
			r.Log.Info("Failed to update status", "err", err)
		}
		return ctrl.Result{}, nil
	}

	// Create the service account and roles
	sa, _ := r.serviceaccount(gateserver)
	err = r.Client.Create(ctx, sa)
	if err != nil {
		r.Log.Info("Failed to create serviceaccount.", "err", err)

		setServerCondition(gateserver, "FailedCreateServiceaccount", err)
		if err := r.Status().Update(ctx, gateserver); err != nil {
			r.Log.Info("Failed to update status", "err", err)
		}
		return ctrl.Result{}, nil
	}
	role, _ := r.role(gateserver)
	err = r.Client.Create(ctx, role)
	if err != nil {
		r.Log.Info("Failed to create role.", "err", err)

		setServerCondition(gateserver, "FailedCreateRole", err)
		if err := r.Status().Update(ctx, gateserver); err != nil {
			r.Log.Info("Failed to update status", "err", err)
		}
		return ctrl.Result{}, nil
	}
	rolebinding, _ := r.rolebinding(gateserver)
	err = r.Client.Create(ctx, rolebinding)
	if err != nil {
		r.Log.Info("Failed to create rolebinding.", "err", err)

		setServerCondition(gateserver, "FailedCreateRolebinding", err)
		if err := r.Status().Update(ctx, gateserver); err != nil {
			r.Log.Info("Failed to update status", "err", err)
		}
		return ctrl.Result{}, nil
	}

	// Create the gate service
	dep, _ := r.deployment(gateserver)
	err = r.Client.Create(ctx, dep)
	if err != nil {
		r.Log.Info("Failed to create deployment.", "err", err)

		setServerCondition(gateserver, "FailedCreateDeployment", err)
		if err := r.Status().Update(ctx, gateserver); err != nil {
			r.Log.Info("Failed to update status", "err", err)
		}
		return ctrl.Result{}, nil
	}

	// Create service and route

	// Add finalizer for this CR
	if !controllerutil.ContainsFinalizer(gateserver, gateserverFinalizer) {
		controllerutil.AddFinalizer(gateserver, gateserverFinalizer)
		err = r.Update(ctx, gateserver)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	gateserver.Status.Phase = "Ready"
	condition := metav1.Condition{
		Type:               "Created",
		Status:             "True",
		Reason:             "AllResourcesCreated",
		Message:            "All resources created",
		LastTransitionTime: t,
	}
	gateserver.Status.Conditions = append(gateserver.Status.Conditions, condition)
	if err := r.Status().Update(ctx, gateserver); err != nil {
		r.Log.Info("Failed to update status", "err", err)
	}

	return ctrl.Result{}, nil
}

func (r *GateServerReconciler) finalizeGateServer(m *ocgatev1beta1.GateServer) error {
	// TODO(user): Add the cleanup steps that the operator
	// needs to do before the CR can be deleted. Examples
	// of finalizers include performing backups and deleting
	// resources that are not owned by this CR, like a PVC.
	r.Log.Info("Successfully finalized gateserver")
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GateServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ocgatev1beta1.GateServer{}).
		Complete(r)
}

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
