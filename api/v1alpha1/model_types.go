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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ModelFinalizer = "sivchari.io/model"
)

// ModelSpec defines the desired state of Model.
type ModelSpec struct {
	// images is a list of images to be used for the ollama. At least one image is required.
	// +required
	// +kubebuilder:validation:MinItems=1
	Images []string `json:"images,omitempty"`

	// paused indicates whether the Model will be provisioned or not.
	// If paused is true, the ollama will not be provisioned.
	// +optional
	Paused *bool `json:"paused,omitempty"`

	// template is the template used to create the ollama server.
	// +optional
	Template *ModelTemplate `json:"template"`
}

type ModelTemplate struct {
	// objectMeta is the metadata used to create the ollama server.
	// +optional
	Metadata *ObjectMeta `json:"metadata,omitempty"`

	// spec is the spec used to create the ollama server.
	// +optional
	Spec *ModelTemplateSpec `json:"spec"`
}

type ModelTemplateSpec struct {
	// volumeMounts is a list of volume mounts to be used for the ollama server.
	// +optional
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`

	// volumes is a list of volumes to be used for the ollama server.
	// +optional
	Volumes []corev1.Volume `json:"volumes,omitempty"`

	// nodeSelector is a selector to restrict the nodes on which the ollama server will be provisioned.
	// +optional
	// +mapType=atomic
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// affinity is a set of rules used to select the nodes on which the ollama server will be provisioned.
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// tolerations is a list of tolerations to be used for the ollama server.
	// +optional
	// +listType=atomic
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// topologySpreadConstraints is a list of topology spread constraints to be used for the ollama server.
	// +optional
	// +patchMergeKey=topologyKey
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=topologyKey
	// +listMapKey=whenUnsatisfiable
	TopologySpreadConstraints []corev1.TopologySpreadConstraint `json:"topologySpreadConstraints,omitempty"`
}

const (
	// ModelConditionAvailable indicates that the model is available, but not need to be ready.
	ModelConditionAvailable = "Available"

	// ModelConditionReady indicates that the model is ready to be used.
	ModelConditionReady = "Ready"

	// ModelConditionFailed indicates that the model has failed.
	ModelConditionFailed = "Failed"
)

// ModelStatus defines the observed state of Model.
type ModelStatus struct {
	// podRef represents a reference to the pod where the model is running.
	// +optional
	PodRef *corev1.ObjectReference `json:"podRef,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Model is the Schema for the models API.
type Model struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ModelSpec   `json:"spec,omitempty"`
	Status ModelStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ModelList contains a list of Model.
type ModelList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Model `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Model{}, &ModelList{})
}
