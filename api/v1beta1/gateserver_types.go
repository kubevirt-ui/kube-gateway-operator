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

// GateServerSpec defines the desired state of GateServer
type GateServerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// apiURL is the k8s API url.
	// Defalut value is "https://kubernetes.default.svc".
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:validation:Pattern="^(http|https)://.*"
	// +kubebuilder:validation:MaxLength=1024
	// +kubebuilder:default:="https://kubernetes.default.svc"
	APIURL string `json:"apiURL,omitempty"`

	// route is the the gate proxy server.
	// +required
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:validation:Pattern="^([a-z0-9-_])+[.]([a-z0-9-_])+[.]([a-z0-9-._])+$"
	// +kubebuilder:validation:MaxLength=226
	Route string `json:"route,omitempty"`

	// serviceAccountNamespace of the rule. "*" represents all namespaces.
	// Defalut value is "*".
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:default:="*"
	ServiceAccountNamespace string `json:"serviceAccountNamespace,omitempty"`

	// serviceAccountVerbs is a list of Verbs that apply to ALL the ResourceKinds and AttributeRestrictions contained in this rule.
	// VerbAll represents all kinds.
	// Defalut value is ["get"].
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="array"
	// +kubebuilder:default:={"get"}
	ServiceAccountVerbs []string `json:"serviceAccountVerbs,omitempty"`

	// serviceAccountAPIGroups is the name of the APIGroup that contains the resources.
	// If multiple API groups are specified, any action requested against one of the enumerated resources in any API group will be allowed.
	// APIGroups is the name of the APIGroup that contains the resources.
	// If multiple API groups are specified, any action requested against one of the enumerated resources in any API group will be allowed.
	// Defalut value is [].
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="array"
	ServiceAccountAPIGroups []string `json:"serviceAccountAPIGroups,omitempty"`

	// serviceAccountResources is a list of resources this rule applies to. '*' represents all resources in the specified apiGroups.
	// '*/foo' represents the subresource 'foo' for all resources in the specified apiGroups.
	// APIGroups is the name of the APIGroup that contains the resources.
	// If multiple API groups are specified, any action requested against one of the enumerated resources in any API group will be allowed.
	// Defalut value is [].
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="array"
	ServiceAccountResources []string `json:"serviceAccountResources,omitempty"`

	// serviceAccountResourceNames is an optional white list of names that the rule applies to.  An empty set means that everything is allowed.
	// APIGroups is the name of the APIGroup that contains the resources.
	// If multiple API groups are specified, any action requested against one of the enumerated resources in any API group will be allowed.
	// Defalut value is [].
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="array"
	ServiceAccountResourceNames []string `json:"serviceAccountResourceNames,omitempty"`

	// serviceAccountNonResourceURLs is a set of partial urls that a user should have access to. *s are allowed, but only as the full, final step in the path
	// If an action is not a resource API request, then the URL is split on '/' and is checked against the NonResourceURLs to look for a match.
	// Since non-resource URLs are not namespaced, this field is only applicable for ClusterRoles referenced from a ClusterRoleBinding.
	// Rules can either apply to API resources (such as "pods" or "secrets") or non-resource URL paths (such as "/api"),  but not both.
	// APIGroups is the name of the APIGroup that contains the resources.
	// If multiple API groups are specified, any action requested against one of the enumerated resources in any API group will be allowed.
	// Defalut value is [].
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="array"
	ServiceAccountNonResourceURLs []string `json:"serviceAccountNonResourceURLs,omitempty"`

	// passThrough determain if the tokens acquired from OAuth2 server directly to k8s API.
	// Defalut value is false.
	// +optional
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="boolean"
	// +kubebuilder:default:=false
	PassThrough bool `json:"passThrough,omitempty"`

	// image is the oc gate proxy image to use.
	// Defalut value is "quay.io/yaacov/kube-gateway:latest".
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:validation:MaxLength=1024
	// +kubebuilder:default:="quay.io/yaacov/kube-gateway:latest"
	Image string `json:"image,omitempty"`

	// webAppImage is the oc gate proxy web application image to use,
	// It's an image including the static web application to be served together
	// with k8s API.
	// The static web application should be in the directory "/data/web/public/"
	// and it will be copied to the proxy servers "/web/public/" directory on pproxy
	// startup. If left empty, the proxies default web application will not be replaced.
	// Defalut value is "" (use default web application).
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:validation:MaxLength=1024
	// +kubebuilder:default:=""
	WebAppImage string `json:"webAppImage,omitempty"`

	// generateSecret determain if a secrete with public and private kes will be automatically
	// generated when the kube-gateway server is created.
	// Defalut value is true.
	// +optional
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="boolean"
	// +kubebuilder:default:=true
	GenerateSecret bool `json:"generateSecret,omitempty"`
}

// GateServerStatus defines the observed state of GateServer
type GateServerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Conditions represent the latest available observations of an object's state
	Conditions []metav1.Condition `json:"conditions"`

	// Token generation phase (ready|error)
	Phase string `json:"phase"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// GateServer is the Schema for the gateservers API
type GateServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GateServerSpec   `json:"spec,omitempty"`
	Status GateServerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GateServerList contains a list of GateServer
type GateServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GateServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GateServer{}, &GateServerList{})
}
