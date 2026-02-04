/*
Copyright 2026.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// JsonServerSpec defines the desired state of JsonServer
type JsonServerSpec struct {
	Replicas   *int32 `json:"replicas,omitempty"`
	JsonConfig string `json:"jsonConfig"`
	Image      string `json:"image,omitempty"`
}

// JsonServerStatus defines the observed state of JsonServer.
type JsonServerStatus struct {
	State   string `json:"state,omitempty"`
	Message string `json:"message,omitempty"`
	// Replicas is the observed number of replicas.
	Replicas int32 `json:"replicas,omitempty"`
	// Selector is the label selector for pods.
	Selector string `json:"selector,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas,selectorpath=.status.selector

// JsonServer is the Schema for the jsonservers API
type JsonServer struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of JsonServer
	// +required
	Spec JsonServerSpec `json:"spec"`

	// status defines the observed state of JsonServer
	// +optional
	Status JsonServerStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// JsonServerList contains a list of JsonServer
type JsonServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []JsonServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&JsonServer{}, &JsonServerList{})
}
