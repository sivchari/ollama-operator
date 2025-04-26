/*
Copyright 2025.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ModelSetFinalizer = "sivchari.io/modelset"
)

// ModelSetSpec defines the desired state of ModelSet.
type ModelSetSpec struct {
	// replicas is the number of replicas in the ModelSet.
	// +optional
	// +kubebuilder:validation:Minimum=1
	Replicas *int32 `json:"replicas,omitempty"`

	// +required
	ModelSpec `json:",inline"`
}

// ModelSetStatus defines the observed state of ModelSet.
type ModelSetStatus struct {
	// replicas is the number of replicas in the ModelSet.
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// availableReplicas is the number of available replicas in the ModelSet.
	// +optional
	AvailableReplicas *int32 `json:"availableReplicas,omitempty"`

	// readyReplicas is the number of ready replicas in the ModelSet.
	// +optional
	ReadyReplicas *int32 `json:"readyReplicas,omitempty"`

	// failedReplicas is the number of desired replicas in the ModelSet.
	// +optional
	FailedReplicas *int32 `json:"failedReplicas,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ModelSet is the Schema for the modelsets API.
type ModelSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ModelSetSpec   `json:"spec,omitempty"`
	Status ModelSetStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ModelSetList contains a list of ModelSet.
type ModelSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ModelSet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ModelSet{}, &ModelSetList{})
}
