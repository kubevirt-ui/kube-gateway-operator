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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GateTokenCache stores initial token data
type GateTokenCache struct {
	From        string `json:"from"`
	Until       string `json:"until"`
	DurationSec int64  `json:"duration-sec"`
	NBf         int64  `json:"nbf"`
	Exp         int64  `json:"exp"`
	MatchMethod string `json:"matchMethod"`
	MatchPath   string `json:"matchPath"`
	Alg         string `json:"alg"`
}

// GateTokenSpec defines the desired state of GateToken
type GateTokenSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// match-path is a regular expresion used to validate API request path,
	// API requests matching this pattern will be validated by the token.
	// This field may not be empty.
	// +required
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:validation:Pattern="^[/^][^:@]+$"
	// +kubebuilder:validation:MaxLength=1024
	MatchPath string `json:"match-path"`

	// from is time of token invocation, the token will not validate before this time,
	// the token duration will start from this time.
	// Defalut to token object creation time.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:validation:Format="date-time"
	From string `json:"from"`

	// duration-sec is the duration in sec the token will be validated since it's invocation.
	// Defalut value is 3600s (1h).
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="integer"
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default:=3600
	DurationSec int64 `json:"duration-sec"`

	// match-path is a comma separated list of allowed http methods,
	// only API requests matching one of the allowed methods will be validated.
	// Defalut value is "GET,OPTIONS".
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:validation:Pattern="^(GET|HEAD|POST|PUT|DELETE|CONNECT|OPTIONS|TRACE)+(,(GET|HEAD|POST|PUT|DELETE|CONNECT|OPTIONS|TRACE)+)*$"
	// +kubebuilder:validation:MaxLength=1024
	// +kubebuilder:default:="GET,OPTIONS"
	MatchMethod string `json:"match-method"`
}

// GateTokenStatus defines the observed state of GateToken
type GateTokenStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Conditions represent the latest available observations of an object's state
	Conditions []metav1.Condition `json:"conditions"`

	// The generated token
	Token string `json:"token"`

	// Cached data, once created, user can not change this valuse
	Data GateTokenCache `json:"data"`

	// Token generation phase (ready|error)
	Phase string `json:"phase"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// GateToken is the Schema for the gatetokens API
type GateToken struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GateTokenSpec   `json:"spec,omitempty"`
	Status GateTokenStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GateTokenList contains a list of GateToken
type GateTokenList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GateToken `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GateToken{}, &GateTokenList{})
}
