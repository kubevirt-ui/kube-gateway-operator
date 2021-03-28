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

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ocgatev1beta1 "github.com/yaacov/oc-gate-operator/api/v1beta1"
)

// GateTokenReconciler reconciles a GateToken object
type GateTokenReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups=security.openshift.io,resources=securitycontextconstraints,resourceNames=privileged,verbs=use
// +kubebuilder:rbac:groups=ocgate.yaacov.com,resources=gatetokens,verbs=get;list;watch;create;delete
// +kubebuilder:rbac:groups=ocgate.yaacov.com,resources=gatetokens/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=ocgate.yaacov.com,resources=gatetokens/finalizers,verbs=update

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
	token := &ocgatev1beta1.GateToken{}
	if err := r.Get(ctx, req.NamespacedName, token); err != nil {
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

	// If token in pending state, check nbf
	if token.Status.Phase == "Pending" {
		now := int64(time.Now().Unix())
		exp := token.Status.Data.Exp
		nbf := token.Status.Data.NBf

		r.Log.Info("Pending phase token", "id", token.Name, "now", now, "nbf", nbf, "exp", exp)

		if (nbf - now) > 0 {
			r.Log.Info("Pending RequeueAfter", "sec", (nbf - now))
			return ctrl.Result{
				RequeueAfter: time.Duration(nbf-now) * time.Second,
			}, nil
		}

		// If token is ready, create sideeffects
		// and set phase to ready

		if token.Spec.GenerateServiceAccount {
			r.Log.Info("Create namespaced service account.")
			sa, _ := r.serviceaccount(token)
			if err := r.Client.Create(ctx, sa); err != nil {
				r.Log.Info("Failed to create serviceaccount", "err", err)
				return ctrl.Result{}, nil
			}

			role, _ := r.role(token)
			if err := r.Client.Create(ctx, role); err != nil {
				r.Log.Info("Failed to create role", "err", err)
				return ctrl.Result{}, nil
			}

			rolebinding, _ := r.rolebinding(token)
			if err := r.Client.Create(ctx, rolebinding); err != nil {
				r.Log.Info("Failed to create rolebinding", "err", err)
				return ctrl.Result{}, nil
			}
		}

		// If token is found, move to Ready
		setReadyCondition(token, "Ready", "Token is ready")

		if err := r.Status().Update(ctx, token); err != nil {
			r.Log.Info("Failed to update status", "err", err)
		}

		// Reques, wating for token.
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	// If token in ready state and missing token, reque
	if token.Status.Phase == "Ready" && token.Status.Token == "" && token.Spec.GenerateServiceAccount {
		sa := &corev1.ServiceAccount{}
		if err := r.Get(ctx, req.NamespacedName, sa); err != nil {
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

		secretPrefix := fmt.Sprintf("%s-token-", token.Name)
		secretName := ""
		for _, s := range sa.Secrets {
			if strings.HasPrefix(s.Name, secretPrefix) {
				secretName = s.Name
			}
		}

		// Reque until token secret is available
		r.Log.Info("Reading token", "secretName", secretName)
		key, err := getSecret(ctx, r.Client, secretName, token.Namespace, "token")
		if err != nil {
			r.Log.Info("Can't read service account token", "err", err)

			setErrorCondition(token, "TokenGetterError", err)
			if err := r.Status().Update(ctx, token); err != nil {
				r.Log.Info("Failed to update status", "err", err)
			}

			return ctrl.Result{}, nil
		}

		token.Status.Token = string(key)
		if err := r.Status().Update(ctx, token); err != nil {
			r.Log.Info("Failed to update status", "err", err)
		}
	}

	// If token in ready state, check expiration
	if token.Status.Phase == "Ready" {
		now := int64(time.Now().Unix())
		exp := token.Status.Data.Exp

		r.Log.Info("Ready phase token", "id", token.Name, "now", now, "exp", exp)

		if (exp - now) > 0 {
			r.Log.Info("Ready RequeueAfter", "sec", (exp - now))
			return ctrl.Result{
				RequeueAfter: time.Duration(exp-now) * time.Second,
			}, nil
		}

		// If token expired, delete sideeffects
		// and set phase to completed

		if token.Spec.GenerateServiceAccount {
			r.Log.Info("Deleting service acount...")

			opts := &client.DeleteOptions{}

			sa := &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      token.Name,
					Namespace: token.Namespace,
				},
			}

			if err := r.Delete(ctx, sa, opts); err != nil {
				r.Log.Info("Failed to delete service account", "err", err)
				return ctrl.Result{}, nil
			}

			role := &rbacv1.Role{
				ObjectMeta: metav1.ObjectMeta{
					Name:      token.Name,
					Namespace: token.Namespace,
				},
			}

			if err := r.Delete(ctx, role, opts); err != nil {
				r.Log.Info("Failed to delete role", "err", err)
				return ctrl.Result{}, nil
			}

			roleBinding := &rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      token.Name,
					Namespace: token.Namespace,
				},
			}
			if err := r.Delete(ctx, roleBinding, opts); err != nil {
				r.Log.Info("Failed to delete roleBinding", "err", err)
				return ctrl.Result{}, nil
			}
		}

		setCompletedCondition(token, "Expired", "Token expired")

		if err := r.Status().Update(ctx, token); err != nil {
			r.Log.Info("Failed to update status", "err", err)
		}
		return ctrl.Result{}, nil
	}

	// If token was created, exit.
	if token.Status.Phase != "" {
		r.Log.Info("Old token", "id", token.Name)
		return ctrl.Result{}, nil
	}

	// Parse and cache user data.
	if err := cacheData(token); err != nil {
		r.Log.Info("Can't parse token data", "err", err)

		setErrorCondition(token, "UserDataError", err)

		if err := r.Status().Update(ctx, token); err != nil {
			r.Log.Info("Failed to update status", "err", err)
		}
		return ctrl.Result{}, nil
	}

	// Set gate-token access code
	// Get private key secret
	if !token.Spec.GenerateServiceAccount {
		key, err := getSecret(ctx, r.Client, "oc-gate-jwt-secret", token.Namespace, "key.pem")
		if err != nil {
			r.Log.Info("Can't read private key secret", "err", err)

			setErrorCondition(token, "PrivateKeyError", err)
			if err := r.Status().Update(ctx, token); err != nil {
				r.Log.Info("Failed to update status", "err", err)
			}

			return ctrl.Result{}, nil
		}

		// Create token
		err = singToken(token, key)
		if err != nil {
			r.Log.Info("Can't read private key secret", "err", err)

			setErrorCondition(token, "PrivateKeyError", err)
			if err := r.Status().Update(ctx, token); err != nil {
				r.Log.Info("Failed to update status", "err", err)
			}

			return ctrl.Result{}, nil
		}
	}

	// Token is ready
	setPendingCondition(token, "TokenCreated", "Token created")
	if err := r.Status().Update(ctx, token); err != nil {
		r.Log.Info("Failed to update status", "err", err)
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GateTokenReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ocgatev1beta1.GateToken{}).
		Complete(r)
}
