package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GitlabSpec defines the desired state of Gitlab
type GitlabSpec struct {
	Replicas    int32               `json:"replicas"`
	Enterprise  bool                `json:"enterprise,omitempty"`
	ExternalURL string              `json:"externalURL,omitempty"`
	Registry    RegistrySpec        `json:"registry,omitempty"`
	Redis       RedisSpec           `json:"redis,omitempty"`
	Database    DatabaseSpec        `json:"database,omitempty"`
	Volumes     ComponentVolumeSpec `json:"volumes,omitempty"`
}

// RedisSpec defines Redis options
type RedisSpec struct {
	Replicas int32 `json:"replicas,omitempty"`
}

// DatabaseSpec defines database options
type DatabaseSpec struct {
	Replicas int32 `json:"replicas,omitempty"`
}

// RegistrySpec defines options for Gitlab registry
type RegistrySpec struct {
	Enabled     bool   `json:"enable,omitempty"`
	ExternalURL string `json:"externalURL,omitempty"`
}

// VolumeSpec defines volume specifications
type VolumeSpec struct {
	// Sets the size of the volume
	Capacity string `json:"capacity,omitempty"`
	// Sets whether the data or volume should persist
	// Should create emptyDir if set to false instead of PVC
	Persist bool `json:"persist,omitempty"`
}

// ComponentVolumeSpec defines volumes for
// the different Gitlab peieces
type ComponentVolumeSpec struct {
	// Postgres database volume for PGDATA
	Postgres VolumeSpec `json:"database,omitempty"`
	// Redis key value store volume
	Redis VolumeSpec `json:"redis,omitempty"`
	// Gitlab configuration volume
	Configuration VolumeSpec `json:"config,omitempty"`
	// Gitlab registry volume
	Registry VolumeSpec `json:"registry,omitempty"`
	// Gitlab rails data volume
	Data VolumeSpec `json:"data,omitempty"`
}

// GitlabStatus defines the observed state of Gitlab
type GitlabStatus struct {
	// Phase represents status of the Gitlab resource
	Phase    string         `json:"phase,omitempty"`
	Stage    string         `json:"stage,omitempty"`
	Services ServicesHealth `json:"services,omitempty"`
}

// ServicesHealth represents the status
//  of a Gitlab dependent service
type ServicesHealth struct {
	Database string `json:"database,omitempty"`
	Redis    string `json:"redis,omitempty"`
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
