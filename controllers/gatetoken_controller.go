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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ocgatev1beta1 "github.com/yaacov/oc-gate-operator/api/v1beta1"
	"github.com/yaacov/oc-gate-operator/pkg/token"
)

const gatetokenFinalizer = "ocgate.yaacov.com/finalizer"

// GateTokenReconciler reconciles a GateToken object
type GateTokenReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups=security.openshift.io,resources=securitycontextconstraints,resourceNames=privileged,verbs=use
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=clusterroles,verbs=get;list;watch;create;update;patch;delete;deletecollection
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=clusterrolebindings,verbs=get;list;watch;create;update;patch;delete;deletecollection
// +kubebuilder:rbac:groups="ocgate.yaacov.com",resources=gatetokens,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="ocgate.yaacov.com",resources=gatetokens/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="ocgate.yaacov.com",resources=gatetokens/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the GateToken object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *GateTokenReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.Info("Reconcile", "gatetoken", req.NamespacedName)

	// your logic here

	// Lookup the GateToken instance for this reconcile request
	t := &ocgatev1beta1.GateToken{}
	if err := r.Get(ctx, req.NamespacedName, t); err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			r.Log.Info("GateToken resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		r.Log.Error(err, "Failed to get GateToken.")
		return ctrl.Result{}, err
	}

	// Check if the GateServer instance is marked to be deleted, which is
	// indicated by the deletion timestamp being set.
	isGateTokenMarkedToBeDeleted := t.GetDeletionTimestamp() != nil
	if isGateTokenMarkedToBeDeleted {
		if controllerutil.ContainsFinalizer(t, gatetokenFinalizer) {
			// Run finalization logic for gatetokenFinalizer. If the
			// finalization logic fails, don't remove the finalizer so
			// that we can retry during the next reconciliation.
			if err := r.finalizeGateToken(t); err != nil {
				return ctrl.Result{}, err
			}

			// Remove gateserverFinalizer. Once all finalizers have been
			// removed, the object will be deleted.
			controllerutil.RemoveFinalizer(t, gateserverFinalizer)
			if err := r.Update(ctx, t); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// If token in pending state, check nbf
	if t.Status.Phase == "Pending" {
		var errs []error
		now := int64(time.Now().Unix())
		exp := t.Status.Data.Exp
		nbf := t.Status.Data.NBf

		r.Log.Info("Pending phase token", "id", t.Name, "now", now, "nbf", nbf, "exp", exp)

		if (nbf - now) > 0 {
			r.Log.Info("Pending RequeueAfter", "sec", (nbf - now))
			return ctrl.Result{
				RequeueAfter: time.Duration(nbf-now) * time.Second,
			}, nil
		}

		// If token is ready, create sideeffects
		// and set phase to ready

		if t.Spec.GenerateServiceAccount {
			r.Log.Info("Create namespaced service account.")
			serviceaccount, _ := token.ServiceAccount(t)
			controllerutil.SetControllerReference(t, serviceaccount, r.Scheme)
			if err := r.Client.Create(ctx, serviceaccount); err != nil {
				r.Log.Info("Failed to create serviceaccount", "err", err)
				errs = append(errs, err)
			}

			clusterrole, _ := token.ClusterRole(t)
			controllerutil.SetControllerReference(t, clusterrole, r.Scheme)
			if err := r.Client.Create(ctx, clusterrole); err != nil {
				r.Log.Info("Failed to create role", "err", err)
				errs = append(errs, err)
			}

			if t.Spec.Namespace == "*" {
				clusterrolebinding, _ := token.ClusterRoleBinding(t)
				controllerutil.SetControllerReference(t, clusterrolebinding, r.Scheme)
				if err := r.Client.Create(ctx, clusterrolebinding); err != nil {
					r.Log.Info("Failed to create clusterrolebinding", "err", err)
					errs = append(errs, err)
				}
			} else {
				rolebinding, _ := token.RoleBinding(t)
				controllerutil.SetControllerReference(t, rolebinding, r.Scheme)
				if err := r.Client.Create(ctx, rolebinding); err != nil {
					r.Log.Info("Failed to create rolebinding", "err", err)
					errs = append(errs, err)
				}
			}
		}

		if len(errs) != 0 {
			setErrorCondition(t, "FailedSA", errs[0])
			if err := r.Status().Update(ctx, t); err != nil {
				r.Log.Info("Failed to update status", "err", err)
			}

			return ctrl.Result{}, nil
		}

		// If token is found, move to Ready
		setReadyCondition(t, "Ready", "Token is ready")
		if err := r.Status().Update(ctx, t); err != nil {
			r.Log.Info("Failed to update status", "err", err)
		}

		// If using k8s to generate the token, reques and wait for token.
		if t.Spec.GenerateServiceAccount {
			return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
		}
	}

	// If token in ready state and missing token, reque
	if t.Status.Phase == "Ready" && t.Status.Token == "" && t.Spec.GenerateServiceAccount {
		serviceaccount := &corev1.ServiceAccount{}
		if err := r.Get(ctx, req.NamespacedName, serviceaccount); err != nil {
			if errors.IsNotFound(err) {
				// Request object not found, could have been deleted after reconcile request.
				// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
				// Return and don't requeue
				r.Log.Info("Service account resource not found. Ignoring since object must be deleted.")
				return ctrl.Result{}, nil
			}
			// Error reading the object - requeue the request.
			r.Log.Error(err, "Failed to get ServiceAccount.")
			return ctrl.Result{}, err
		}

		secretPrefix := fmt.Sprintf("%s-token-", t.Name)
		secretName := ""
		for _, s := range serviceaccount.Secrets {
			if strings.HasPrefix(s.Name, secretPrefix) {
				secretName = s.Name
			}
		}

		// Reque until token secret is available
		r.Log.Info("Reading token", "secretName", secretName)
		key, err := getSecret(ctx, r.Client, secretName, t.Namespace, "token")
		if err != nil {
			r.Log.Info("Can't read service account token", "err", err)

			setErrorCondition(t, "TokenGetterError", err)
			if err := r.Status().Update(ctx, t); err != nil {
				r.Log.Info("Failed to update status", "err", err)
			}

			return ctrl.Result{}, nil
		}

		t.Status.Token = string(key)
		if err := r.Status().Update(ctx, t); err != nil {
			r.Log.Info("Failed to update status", "err", err)
		}
	}

	// If token in ready state, check expiration
	if t.Status.Phase == "Ready" {
		now := int64(time.Now().Unix())
		exp := t.Status.Data.Exp

		r.Log.Info("Ready phase token", "id", t.Name, "now", now, "exp", exp)

		if (exp - now) > 0 {
			r.Log.Info("Ready RequeueAfter", "sec", (exp - now))
			return ctrl.Result{
				RequeueAfter: time.Duration(exp-now) * time.Second,
			}, nil
		}

		// If token expired, delete sideeffects
		// and set phase to completed

		if t.Spec.GenerateServiceAccount {
			r.Log.Info("Deleting service acount...")

			opts := &client.DeleteOptions{}
			errs := []error{}

			serviceaccount := &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      t.Name,
					Namespace: t.Namespace,
				},
			}
			if err := r.Delete(ctx, serviceaccount, opts); err != nil {
				r.Log.Info("Failed to delete service account", "err", err)
				errs = append(errs, err)
			}

			clusterrole := &rbacv1.ClusterRole{
				ObjectMeta: metav1.ObjectMeta{
					Name: t.Name,
				},
			}
			if err := r.Delete(ctx, clusterrole, opts); err != nil {
				r.Log.Info("Failed to delete role", "err", err)
				errs = append(errs, err)
			}

			if t.Spec.Namespace == "*" {
				clusterrolebinding := &rbacv1.ClusterRoleBinding{
					ObjectMeta: metav1.ObjectMeta{
						Name: t.Name,
					},
				}
				if err := r.Delete(ctx, clusterrolebinding, opts); err != nil {
					r.Log.Info("Failed to delete clusterRoleBinding", "err", err)
					errs = append(errs, err)
				}
			} else {
				rolebinding := &rbacv1.RoleBinding{
					ObjectMeta: metav1.ObjectMeta{
						Name:      t.Name,
						Namespace: t.Spec.Namespace,
					},
				}
				if err := r.Delete(ctx, rolebinding, opts); err != nil {
					r.Log.Info("Failed to delete roleBinding", "err", err)
					errs = append(errs, err)
				}
			}

			if len(errs) != 0 {
				r.Log.Info("Failed to delete service account")
			}
		}

		setCompletedCondition(t, "Expired", "Token expired")

		if err := r.Status().Update(ctx, t); err != nil {
			r.Log.Info("Failed to update status", "err", err)
		}
		return ctrl.Result{}, nil
	}

	// If token was created, exit.
	if t.Status.Phase != "" {
		r.Log.Info("Old token", "id", t.Name)
		return ctrl.Result{}, nil
	}

	// Check role
	var err error
	if len(t.Spec.NonResourceURLs) != 0 && len(t.Spec.APIGroups) != 0 {
		err = fmt.Errorf("auth roles can either apply to API resources or non-resource URL paths, but not both")
	}
	if len(t.Spec.NonResourceURLs) == 0 && len(t.Spec.APIGroups) == 0 {
		err = fmt.Errorf("auth roles can either apply to API resources or non-resource URL paths, but can't be empty")
	}

	if err != nil {
		r.Log.Info("Failed to create oc gate token.", "err", err)

		setErrorCondition(t, "UserDataError", err)
		if err := r.Status().Update(ctx, t); err != nil {
			r.Log.Info("Failed to update status", "err", err)
		}
		return ctrl.Result{}, nil
	}

	// Parse and cache user data.
	if err := cacheData(t); err != nil {
		r.Log.Info("Can't parse token data", "err", err)

		setErrorCondition(t, "UserDataError", err)
		if err := r.Status().Update(ctx, t); err != nil {
			r.Log.Info("Failed to update status", "err", err)
		}
		return ctrl.Result{}, nil
	}

	// Set gate-token access code
	// Get private key secret
	if !t.Spec.GenerateServiceAccount {
		key, err := getSecret(ctx, r.Client, "oc-gate-jwt-secret", t.Namespace, "key.pem")
		if err != nil {
			r.Log.Info("Can't read private key secret", "err", err)

			setErrorCondition(t, "PrivateKeyError", err)
			if err := r.Status().Update(ctx, t); err != nil {
				r.Log.Info("Failed to update status", "err", err)
			}
			return ctrl.Result{}, nil
		}

		// Create token
		err = singToken(t, key)
		if err != nil {
			r.Log.Info("Can't read private key secret", "err", err)

			setErrorCondition(t, "PrivateKeyError", err)
			if err := r.Status().Update(ctx, t); err != nil {
				r.Log.Info("Failed to update status", "err", err)
			}

			return ctrl.Result{}, nil
		}
	}

	// Add finalizer for this CR
	if !controllerutil.ContainsFinalizer(t, gatetokenFinalizer) {
		controllerutil.AddFinalizer(t, gatetokenFinalizer)
		if err := r.Update(ctx, t); err != nil {
			r.Log.Info("Failed to add finalizer", "err", err)
			return ctrl.Result{}, nil
		}
	}

	// Token is ready
	setPendingCondition(t, "TokenPending", "Token pending")
	if err := r.Status().Update(ctx, t); err != nil {
		r.Log.Info("Failed to update status", "err", err)
	}
	return ctrl.Result{}, nil
}

func (r *GateTokenReconciler) finalizeGateToken(t *ocgatev1beta1.GateToken) error {
	// TODO(user): Add the cleanup steps that the operator
	// needs to do before the CR can be deleted. Examples
	// of finalizers include performing backups and deleting
	// resources that are not owned by this CR, like a PVC.

	ctx := context.Background()
	opts := &client.DeleteOptions{}
	errs := []error{}

	if !t.Spec.GenerateServiceAccount {
		r.Log.Info("Successfully finalized gatetoken (no ServiceAccount)")
		return nil
	}

	r.Log.Info("Deleting cluster role and cluster role binding...")

	clusterrole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: t.Name,
		},
	}
	if err := r.Delete(ctx, clusterrole, opts); err != nil {
		r.Log.Info("Failed to finalize clusterRole", "err", err)
		errs = append(errs, err)
	}

	if t.Spec.Namespace == "*" {
		clusterrolebinding := &rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: t.Name,
			},
		}
		if err := r.Delete(ctx, clusterrolebinding, opts); err != nil {
			r.Log.Info("Failed to finalize clusterRoleBinding", "err", err)
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 {
		r.Log.Info("Failed to finalized gatetoken")
	} else {
		r.Log.Info("Successfully finalized gatetoken (ServiceAccount)")
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GateTokenReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ocgatev1beta1.GateToken{}).
		Complete(r)
}
