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
	"time"

	"github.com/go-logr/logr"
	"github.com/golang-jwt/jwt"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kubegatewayv1beta1 "github.com/kubevirt-ui/kube-gateway-operator/api/v1beta1"
)

// GateTokenReconciler reconciles a GateToken object
type GateTokenReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups=security.openshift.io,resources=securitycontextconstraints,resourceNames=privileged,verbs=use
// +kubebuilder:rbac:groups=kubegateway.kubevirt.io,resources=gatetokens,verbs=get;list;watch;create;delete
// +kubebuilder:rbac:groups=kubegateway.kubevirt.io,resources=gatetokens/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=kubegateway.kubevirt.io,resources=gatetokens/finalizers,verbs=update

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
	_ = r.Log.WithValues("gatetoken", req.NamespacedName)

	// your logic here

	// Lookup the GateToken instance for this reconcile request
	token := &kubegatewayv1beta1.GateToken{}
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

	// Get private key secret
	key, err := getSecret(ctx, r.Client, token.Spec.SecretName, token.Spec.SecretNamespace, token.Spec.SecretFile)
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

	// Token is ready
	setReadyCondition(token, "TokenCreated", "token created")
	if err := r.Status().Update(ctx, token); err != nil {
		r.Log.Info("Failed to update status", "err", err)
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GateTokenReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kubegatewayv1beta1.GateToken{}).
		Complete(r)
}

// Cache user data
func cacheData(token *kubegatewayv1beta1.GateToken) error {
	var notBeforeTime int64
	var duration time.Duration

	// Default from is "now"
	if token.Spec.From == "" {
		notBeforeTime = int64(time.Now().Unix())
	} else {
		fromTime, err := time.Parse(time.RFC3339, token.Spec.From)
		if err != nil {
			return err
		}
		notBeforeTime = int64(fromTime.Unix())
	}

	// Default Verbs is ["get"]
	if token.Spec.Verbs == nil {
		token.Spec.Verbs = []string{"get"}
	}

	// Default DurationSec is 3600s (1h)
	if token.Spec.Duration == "" {
		token.Spec.Duration = "1h"
	}

	if token.Spec.SecretNamespace == "" {
		token.Spec.SecretNamespace = token.Namespace
	}

	// Set gate token cache data
	duration, _ = time.ParseDuration(token.Spec.Duration)
	token.Status.Data.NBf = notBeforeTime
	token.Status.Data.Exp = notBeforeTime + int64(duration.Seconds())
	token.Status.Data.From = time.Unix(notBeforeTime, 0).UTC().Format(time.RFC3339)
	token.Status.Data.Until = time.Unix(notBeforeTime+int64(duration.Seconds()), 0).UTC().Format(time.RFC3339)
	token.Status.Data.Duration = token.Spec.Duration
	token.Status.Data.Verbs = token.Spec.Verbs
	token.Status.Data.URLs = token.Spec.URLs

	return nil
}

func getSecret(ctx context.Context, client client.Client, name string, namespace string, file string) ([]byte, error) {
	// Get private key secret
	secret := &corev1.Secret{}
	namespaced := &types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}

	if err := client.Get(ctx, *namespaced, secret); err != nil {
		return nil, err
	}

	key := secret.Data[file]
	return key, nil
}

func setErrorCondition(token *kubegatewayv1beta1.GateToken, reason string, err error) {
	t := metav1.Time{Time: time.Now()}
	token.Status.Phase = "Error"
	condition := metav1.Condition{
		Type:               "Error",
		Status:             "True",
		Reason:             reason,
		Message:            fmt.Sprintf("%s", err),
		LastTransitionTime: t,
	}
	token.Status.Conditions = []metav1.Condition{condition}
}

func setReadyCondition(token *kubegatewayv1beta1.GateToken, reason string, message string) {
	t := metav1.Time{Time: time.Now()}
	token.Status.Phase = "Ready"
	condition := metav1.Condition{
		Type:               "Ready",
		Status:             "True",
		Reason:             reason,
		Message:            message,
		LastTransitionTime: t,
	}
	token.Status.Conditions = []metav1.Condition{condition}
}

func singToken(token *kubegatewayv1beta1.GateToken, key []byte) error {
	// Create token
	claims := &jwt.MapClaims{
		"exp":   token.Status.Data.Exp,
		"nbf":   token.Status.Data.NBf,
		"URLs":  token.Status.Data.URLs,
		"verbs": token.Status.Data.Verbs,
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	jwtKey, err := jwt.ParseRSAPrivateKeyFromPEM(key)
	if err != nil {
		return err
	}
	out, err := jwtToken.SignedString(jwtKey)
	if err != nil {
		return err
	}

	token.Status.Token = out
	return nil
}
