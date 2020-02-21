package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GitlabSpec defines the desired state of Gitlab
type GitlabSpec struct {
	Replicas      int32        `json:"replicas"`
	Enterprise    bool         `json:"enterprise,omitempty"`
	ExternalURL   string       `json:"externalURL,omitempty"`
	Configuration ConfigSpec   `json:"config,omitempty"`
	Volume        VolumeSpec   `json:"volume,omitempty"`
	Database      DatabaseSpec `json:"database,omitempty"`
	Redis         RedisSpec    `json:"redis,omitempty"`
	Runner        RunnerSpec   `json:"runner,omitempty"`
	Registry      RegistrySpec `json:"registry,omitempty"`
}

// RedisSpec defines Redis options
type RedisSpec struct {
	Replicas int32      `json:"replicas,omitempty"`
	Volume   VolumeSpec `json:"volume,omitempty"`
}

// ConfigSpec defines Redis options
type ConfigSpec struct {
	Volume VolumeSpec `json:"volume,omitempty"`
}

// DatabaseSpec defines database options
type DatabaseSpec struct {
	Replicas int32      `json:"replicas,omitempty"`
	Volume   VolumeSpec `json:"volume,omitempty"`
}

// RunnerSpec defines options for Gitlab runner
type RunnerSpec struct {
	Replicas int32 `json:"replicas,omitempty"`
	Enabled  bool  `json:"enable,omitempty"`
}

// RegistrySpec defines options for Gitlab registry
type RegistrySpec struct {
	Replicas int32      `json:"replicas,omitempty"`
	Enabled  bool       `json:"enable,omitempty"`
	Volume   VolumeSpec `json:"volume,omitempty"`
}

// VolumeSpec defines volume specifications
type VolumeSpec struct {
	// Sets the size of the volume in Gi
	Capacity string `json:"capacity,omitempty"`
	// Sets whether the data or volume should persist
	// Should create emptyDir if set to false instead of PVC
	Persist bool `json:"persist,omitempty"`
}

// GitlabStatus defines the observed state of Gitlab
type GitlabStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Gitlab is the Schema for the gitlabs API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=gitlabs,scope=Namespaced
type Gitlab struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GitlabSpec   `json:"spec,omitempty"`
	Status GitlabStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GitlabList contains a list of Gitlab
type GitlabList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Gitlab `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Gitlab{}, &GitlabList{})
}
