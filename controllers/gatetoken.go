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

	ocgatev1beta1 "github.com/yaacov/oc-gate-operator/api/v1beta1"
)

// Cache user data
func cacheData(token *ocgatev1beta1.GateToken) error {
	var nbf int64

	if token.Spec.From == "" {
		nbf = int64(time.Now().Unix())
	} else {
		t, err := time.Parse(time.RFC3339, token.Spec.From)
		if err != nil {
			return err
		}
		nbf = int64(t.Unix())
	}

	token.Status.Data = ocgatev1beta1.GateTokenCache{
		NBf:         nbf,
		Exp:         nbf + int64(token.Spec.DurationSec),
		From:        token.Spec.From,
		Until:       time.Unix(nbf+int64(token.Spec.DurationSec), 0).UTC().Format(time.RFC3339),
		DurationSec: token.Spec.DurationSec,
		MatchMethod: token.Spec.MatchMethod,
		MatchPath:   token.Spec.MatchPath,
		Alg:         jwt.SigningMethodRS256.Name,
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

func setErrorCondition(token *ocgatev1beta1.GateToken, reason string, err error) {
	t := metav1.Time{Time: time.Now()}
	token.Status.Phase = "Error"
	condition := metav1.Condition{
		Type:               "Error",
		Status:             "True",
		Reason:             reason,
		Message:            fmt.Sprintf("%s", err),
		LastTransitionTime: t,
	}
	token.Status.Conditions = append(token.Status.Conditions, condition)
}

func setReadyCondition(token *ocgatev1beta1.GateToken, reason string, message string) {
	t := metav1.Time{Time: time.Now()}
	token.Status.Phase = "Ready"
	condition := metav1.Condition{
		Type:               "Ready",
		Status:             "True",
		Reason:             reason,
		Message:            message,
		LastTransitionTime: t,
	}
	token.Status.Conditions = append(token.Status.Conditions, condition)
}

func setPendingCondition(token *ocgatev1beta1.GateToken, reason string, message string) {
	t := metav1.Time{Time: time.Now()}
	token.Status.Phase = "Pending"
	condition := metav1.Condition{
		Type:               "Pending",
		Status:             "True",
		Reason:             reason,
		Message:            message,
		LastTransitionTime: t,
	}
	token.Status.Conditions = append(token.Status.Conditions, condition)
}

func setCompletedCondition(token *ocgatev1beta1.GateToken, reason string, message string) {
	t := metav1.Time{Time: time.Now()}
	token.Status.Phase = "Completed"
	condition := metav1.Condition{
		Type:               "Completed",
		Status:             "True",
		Reason:             reason,
		Message:            message,
		LastTransitionTime: t,
	}
	token.Status.Conditions = append(token.Status.Conditions, condition)
}

func singToken(token *ocgatev1beta1.GateToken, key []byte) error {
	// Create token
	claims := &jwt.MapClaims{
		"exp":         token.Status.Data.Exp,
		"nbf":         token.Status.Data.NBf,
		"matchPath":   token.Status.Data.MatchPath,
		"matchMethod": token.Status.Data.MatchMethod,
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
