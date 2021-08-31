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

	// img is the kube-gateway image to use.
	// Defalut value is "quay.io/kubevirt-ui/kube-gateway:latest".
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:validation:MaxLength=1024
	// +kubebuilder:default:="quay.io/kubevirt-ui/kube-gateway:latest"
	IMG string `json:"img,omitempty"`

	// api-url is the k8s API url.
	// Defalut value is "https://kubernetes.default.svc".
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:validation:Pattern="^(http|https)://.*"
	// +kubebuilder:validation:MaxLength=1024
	// +kubebuilder:default:="https://kubernetes.default.svc"
	APIURL string `json:"api-url,omitempty"`

	// route for the gate proxy server.
	// +required
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:validation:Pattern="^([a-z0-9-_])+[.]([a-z0-9-_])+[.]([a-z0-9-._])+$"
	// +kubebuilder:validation:MaxLength=226
	Route string `json:"route,omitempty"`

	// admin-role is the verbs athorization role of the service (reader/admin)
	// if service is role is reader, clients getting tokens to use this service
	// will be able to excute get, watch and list verbs.
	// if service is role is admin, clients getting tokens to use this service
	// will be able to excute get, watch, list, patch, creat and delete verbs.
	// Defalut value is "reader".
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:validation:Pattern="^(reader|admin)$"
	// +kubebuilder:validation:MaxLength=1024
	// +kubebuilder:default:="reader"
	AdminRole string `json:"admin-role,omitempty"`

	// admin-resources is a comma separated list of resources athorization role of the service
	// if left empty service could access any resource.
	// Defalut value is "".
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type="string"
	// +kubebuilder:validation:MaxLength=1024
	// +kubebuilder:default:=""
	AdminResources string `json:"admin-resources,omitempty"`
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
