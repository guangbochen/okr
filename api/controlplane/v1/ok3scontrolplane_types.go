/*
Copyright 2023.

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

const (
	// Ok3sControlPlaneFinalizer allows the controller to clean up resources on delete.
	Ok3sControlPlaneFinalizer = "ok3s.controlplane.cluster.x-k8s.io"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// Ok3sControlPlaneSpec defines the desired state of Ok3sControlPlane
type Ok3sControlPlaneSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Ok3sControlPlane. Edit ok3scontrolplane_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// Ok3sControlPlaneStatus defines the observed state of Ok3sControlPlane
type Ok3sControlPlaneStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Ok3sControlPlane is the Schema for the ok3scontrolplanes API
type Ok3sControlPlane struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   Ok3sControlPlaneSpec   `json:"spec,omitempty"`
	Status Ok3sControlPlaneStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// Ok3sControlPlaneList contains a list of Ok3sControlPlane
type Ok3sControlPlaneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Ok3sControlPlane `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Ok3sControlPlane{}, &Ok3sControlPlaneList{})
}
