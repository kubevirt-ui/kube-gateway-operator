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

	"github.com/dgrijalva/jwt-go"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ocgatev1beta1 "github.com/yaacov/virt-gateway-operator/api/v1beta1"
)

// Cache user data
func cacheData(t *ocgatev1beta1.GateToken) error {
	var notBeforeTime int64

	if t.Spec.From == "" {
		notBeforeTime = int64(time.Now().Unix())
		t.Spec.From = time.Unix(notBeforeTime, 0).UTC().Format(time.RFC3339)
	} else {
		fromTime, err := time.Parse(time.RFC3339, t.Spec.From)
		if err != nil {
			return err
		}
		notBeforeTime = int64(fromTime.Unix())
	}

	t.Status.Data = ocgatev1beta1.GateTokenCache{
		From:            t.Spec.From,
		Until:           time.Unix(notBeforeTime+int64(t.Spec.DurationSec), 0).UTC().Format(time.RFC3339),
		DurationSec:     t.Spec.DurationSec,
		NBf:             notBeforeTime,
		Exp:             notBeforeTime + int64(t.Spec.DurationSec),
		Alg:             jwt.SigningMethodRS256.Name,
		Namespace:       t.Spec.Namespace,
		Verbs:           t.Spec.Verbs,
		APIGroups:       t.Spec.APIGroups,
		Resources:       t.Spec.Resources,
		ResourceNames:   t.Spec.ResourceNames,
		NonResourceURLs: t.Spec.NonResourceURLs,
	}

	return nil
}

func getSecret(ctx context.Context, client client.Client, name string, namespace string, secretName string) ([]byte, error) {
	// Get private key secret
	secret := &corev1.Secret{}
	namespaced := &types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}

	if err := client.Get(ctx, *namespaced, secret); err != nil {
		return nil, err
	}

	key := secret.Data[secretName]
	return key, nil
}

func setErrorCondition(t *ocgatev1beta1.GateToken, reason string, err error) {
	now := metav1.Time{Time: time.Now()}
	t.Status.Phase = "Error"
	condition := metav1.Condition{
		Type:               "Error",
		Status:             "True",
		Reason:             reason,
		Message:            fmt.Sprintf("%s", err),
		LastTransitionTime: now,
	}
	t.Status.Conditions = append(t.Status.Conditions, condition)
}

func setReadyCondition(t *ocgatev1beta1.GateToken, reason string, message string) {
	now := metav1.Time{Time: time.Now()}
	t.Status.Phase = "Ready"
	condition := metav1.Condition{
		Type:               "Ready",
		Status:             "True",
		Reason:             reason,
		Message:            message,
		LastTransitionTime: now,
	}
	t.Status.Conditions = append(t.Status.Conditions, condition)
}

func setPendingCondition(t *ocgatev1beta1.GateToken, reason string, message string) {
	now := metav1.Time{Time: time.Now()}
	t.Status.Phase = "Pending"
	condition := metav1.Condition{
		Type:               "Pending",
		Status:             "True",
		Reason:             reason,
		Message:            message,
		LastTransitionTime: now,
	}
	t.Status.Conditions = append(t.Status.Conditions, condition)
}

func setCompletedCondition(t *ocgatev1beta1.GateToken, reason string, message string) {
	now := metav1.Time{Time: time.Now()}
	t.Status.Phase = "Completed"
	condition := metav1.Condition{
		Type:               "Completed",
		Status:             "True",
		Reason:             reason,
		Message:            message,
		LastTransitionTime: now,
	}
	t.Status.Conditions = append(t.Status.Conditions, condition)
}

func singToken(t *ocgatev1beta1.GateToken, key []byte) error {
	// Create token
	claims := &jwt.MapClaims{
		"exp":             t.Status.Data.Exp,
		"nbf":             t.Status.Data.NBf,
		"namespace":       t.Status.Data.Namespace,
		"verbs":           t.Status.Data.Verbs,
		"apiGroups":       t.Status.Data.APIGroups,
		"resources":       t.Status.Data.Resources,
		"resourceNames":   t.Status.Data.ResourceNames,
		"nonResourceURLs": t.Status.Data.NonResourceURLs,
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

	t.Status.Token = out
	return nil
}
