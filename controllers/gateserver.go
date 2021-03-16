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

	ocgatev1beta1 "github.com/yaacov/oc-gate-operator/api/v1beta1"
)

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

func (r *GateServerReconciler) buildServer(ctx context.Context, s *ocgatev1beta1.GateServer) error {
	// Create the service and route
	r.Log.Info("Create the service and route.")
	se, _ := r.service(s)
	err := r.Client.Create(ctx, se)
	if err != nil {
		return err
	}

	route, _ := r.route(s)
	err = r.Client.Create(ctx, route)
	if err != nil {
		return err
	}

	// Create the service account and roles
	r.Log.Info("Create the service account and roles.")
	sa, _ := r.serviceaccount(s)
	err = r.Client.Create(ctx, sa)
	if err != nil {
		return err
	}
	if s.Spec.AdminNamespaced {
		r.Log.Info("Create namespaced role.")
		role, _ := r.role(s)
		err = r.Client.Create(ctx, role)
		if err != nil {
			return err
		}

		rolebinding, _ := r.rolebinding(s)
		err = r.Client.Create(ctx, rolebinding)
		if err != nil {
			return err
		}
	} else {
		r.Log.Info("Create cluster role.")
		role, _ := r.clusterrole(s)
		err = r.Client.Create(ctx, role)
		if err != nil {
			return err
		}

		rolebinding, _ := r.clusterrolebinding(s)
		err = r.Client.Create(ctx, rolebinding)
		if err != nil {
			return err
		}
	}

	// Create the JWT secret
	if s.Spec.GenerateSecret {
		r.Log.Info("Create JWT secret.")
		secret, _ := r.secret(s)
		err = r.Client.Create(ctx, secret)
		if err != nil {
			return err
		}
	}

	// Create the gate service
	r.Log.Info("Create deployment.")
	dep, _ := r.deployment(s)
	err = r.Client.Create(ctx, dep)
	if err != nil {
		return err
	}

	// Create the oauthclient
	r.Log.Info("Create oauthclient.")
	oauth, _ := r.oauthclient(s)
	err = r.Client.Create(ctx, oauth)
	if err != nil {
		return err
	}

	return nil
}
