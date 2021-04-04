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

package proxy

import (
	"fmt"

	oauthv1 "github.com/openshift/api/oauth/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ocgatev1beta1 "github.com/yaacov/virt-gateway-operator/api/v1beta1"
)

// OAuthClient is a
func OAuthClient(s *ocgatev1beta1.GateServer) (*oauthv1.OAuthClient, error) {
	labels := map[string]string{
		"app": s.Name,
	}

	oauthclient := &oauthv1.OAuthClient{
		ObjectMeta: metav1.ObjectMeta{
			Name:   s.Name,
			Labels: labels,
		},
		GrantMethod:  oauthv1.GrantHandlerAuto,
		Secret:       fmt.Sprintf("%s-oauth-secret", s.Name),
		RedirectURIs: []string{fmt.Sprintf("https://%s/auth/callback", s.Spec.Route)},
	}

	return oauthclient, nil
}
