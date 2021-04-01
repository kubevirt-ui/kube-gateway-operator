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
	From            string   `json:"from"`
	Until           string   `json:"until"`
	DurationSec     int64    `json:"duration-sec"`
	NBf             int64    `json:"nbf"`
	Exp             int64    `json:"exp"`
	Alg             string   `json:"alg"`
	Verbs           []string `json:"verbs,omitempty"`
	APIGroups       []string `json:"APIGroups,omitempty"`
	Resources       []string `json:"resources,omitempty"`
	ResourceNames   []string `json:"resourceNames,omitempty"`
	NonResourceURLs []string `json:"nonResourceURLs,omitempty"`
}

// GateTokenSpec defines the desired state of GateToken
type GateTokenSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// from is time of token invocation, the token will not validate before this time,
	// the token duration will start from this time.
	// Defalut to token object creation time.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:validation:Format="date-time"
	From string `json:"from"`

	// durationSec is the duration in sec the token will be validated since it's invocation.
	// Defalut value is 3600s (1h).
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="integer"
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default:=3600
	DurationSec int64 `json:"durationSec"`

	// generateServiceAccount determain if the operator will create a service account and
	// delever the actual service account token instead of a JWT access key.
	// the service account will be generated not before the token is valid
	// and will be deleted when the token expires.
	// Defalut value is false.
	// +optional
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="boolean"
	// +kubebuilder:default:=false
	GenerateServiceAccount bool `json:"generateServiceAccount,omitempty"`

	// verbs is a list of Verbs that apply to ALL the ResourceKinds and AttributeRestrictions contained in this rule.  VerbAll represents all kinds.
	// APIGroups is the name of the APIGroup that contains the resources.
	// If multiple API groups are specified, any action requested against one of the enumerated resources in any API group will be allowed.
	// Defalut value is ["get"].
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="array"
	// +kubebuilder:default:={"get"}
	Verbs []string `json:"verbs,omitempty"`

	// APIGroups is the name of the APIGroup that contains the resources.
	// If multiple API groups are specified, any action requested against one of the enumerated resources in any API group will be allowed.
	// Defalut value is [].
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="array"
	APIGroups []string `json:"APIGroups,omitempty"`

	// resources is a list of resources this rule applies to.  '*' represents all resources in the specified apiGroups.
	// '*/foo' represents the subresource 'foo' for all resources in the specified apiGroups.
	// APIGroups is the name of the APIGroup that contains the resources.
	// If multiple API groups are specified, any action requested against one of the enumerated resources in any API group will be allowed.
	// Defalut value is [].
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="array"
	Resources []string `json:"resources,omitempty"`

	// resourceNames is an optional white list of names that the rule applies to.  An empty set means that everything is allowed.
	// APIGroups is the name of the APIGroup that contains the resources.
	// If multiple API groups are specified, any action requested against one of the enumerated resources in any API group will be allowed.
	// Defalut value is [].
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="array"
	ResourceNames []string `json:"resourceNames,omitempty"`

	// nonResourceURLs is a set of partial urls that a user should have access to.  *s are allowed, but only as the full, final step in the path
	// If an action is not a resource API request, then the URL is split on '/' and is checked against the NonResourceURLs to look for a match.
	// Since non-resource URLs are not namespaced, this field is only applicable for ClusterRoles referenced from a ClusterRoleBinding.
	// Rules can either apply to API resources (such as "pods" or "secrets") or non-resource URL paths (such as "/api"),  but not both.
	// APIGroups is the name of the APIGroup that contains the resources.
	// If multiple API groups are specified, any action requested against one of the enumerated resources in any API group will be allowed.
	// Defalut value is [].
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="array"
	NonResourceURLs []string `json:"nonResourceURLs,omitempty"`
}

// GateTokenStatus defines the observed state of GateToken
type GateTokenStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Conditions represent the latest available observations of an object's state
	Conditions []metav1.Condition `json:"conditions"`

	// The generated token
	Token string `json:"token"`

	// The generated service account name
	ServiceAccountName string `json:"service-account-name"`

	// Cached data, once created, user can not change this valuse
	Data GateTokenCache `json:"data"`

	// Token generation phase (pending|ready|expired|error)
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
