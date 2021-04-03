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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	ocgatev1beta1 "github.com/yaacov/oc-gate-operator/api/v1beta1"
	"github.com/yaacov/oc-gate-operator/pkg/proxy"
)

func setServerCondition(s *ocgatev1beta1.GateServer, reason string, err error) {
	now := metav1.Time{Time: time.Now()}
	s.Status.Phase = "Error"
	condition := metav1.Condition{
		Type:               "Error",
		Status:             "True",
		Reason:             reason,
		Message:            fmt.Sprintf("%s", err),
		LastTransitionTime: now,
	}
	s.Status.Conditions = []metav1.Condition{condition}
}

func (r *GateServerReconciler) buildServer(ctx context.Context, s *ocgatev1beta1.GateServer) error {
	// Create the service and route
	r.Log.Info("Create the service and route.")
	service, _ := proxy.Service(s)
	controllerutil.SetControllerReference(s, service, r.Scheme)
	if err := r.Client.Create(ctx, service); err != nil {
		return err
	}

	route, _ := proxy.Route(s)
	controllerutil.SetControllerReference(s, route, r.Scheme)
	if err := r.Client.Create(ctx, route); err != nil {
		return err
	}

	// Create the service account and roles
	r.Log.Info("Create the service account and roles.")
	serviceaccount, _ := proxy.ServiceAccount(s)
	controllerutil.SetControllerReference(s, serviceaccount, r.Scheme)
	if err := r.Client.Create(ctx, serviceaccount); err != nil {
		return err
	}

	r.Log.Info("Create cluster role.")
	role, _ := proxy.ClusterRole(s)
	controllerutil.SetControllerReference(s, role, r.Scheme)
	if err := r.Client.Create(ctx, role); err != nil {
		return err
	}

	if s.Spec.ServiceAccountNamespace == "*" {
		clusterrolebinding, _ := proxy.ClusterRoleBinding(s)
		controllerutil.SetControllerReference(s, clusterrolebinding, r.Scheme)
		if err := r.Client.Create(ctx, clusterrolebinding); err != nil {
			return err
		}
	} else {
		rolebinding, _ := proxy.RoleBinding(s)
		controllerutil.SetControllerReference(s, rolebinding, r.Scheme)
		if err := r.Client.Create(ctx, rolebinding); err != nil {
			return err
		}
	}

	// Create the JWT secret
	if s.Spec.GenerateSecret {
		r.Log.Info("Create JWT secret.")
		secret, _ := proxy.Secret(s)
		controllerutil.SetControllerReference(s, secret, r.Scheme)
		if err := r.Client.Create(ctx, secret); err != nil {
			return err
		}
	}

	// Create the gate service
	r.Log.Info("Create deployment.")
	deployment, _ := proxy.Deployment(s)
	controllerutil.SetControllerReference(s, deployment, r.Scheme)
	if err := r.Client.Create(ctx, deployment); err != nil {
		return err
	}

	// Create the oauthclient
	r.Log.Info("Create oauthclient.")
	oauthclient, _ := proxy.OAuthClient(s)
	controllerutil.SetControllerReference(s, oauthclient, r.Scheme)
	if err := r.Client.Create(ctx, oauthclient); err != nil {
		return err
	}

	return nil
}
