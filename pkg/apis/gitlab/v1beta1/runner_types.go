package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RunnerSpec defines the desired state of Runner
type RunnerSpec struct {
	// GitlabResource represents a Gitlab custom resource. Should
	// only be used to reference Gitlab instance created by the operator
	Gitlab GitlabInstanceSpec `json:"gitlab,omitempty"`
}

// GitlabInstanceSpec defines the Gitlab custom
// resource in the kubernetes
type GitlabInstanceSpec struct {
	// Name of gitlab resource in kubernetes / openshift
	Name string `json:"name,omitempty"`
	// Gitlab or Continuous Integration URL
	URL string `json:"url,omitempty"`
	// Gitlab Runner registration token
	RegistrationToken string `json:"token,omitempty"`
}

// RunnerStatus defines the observed state of Runner
type RunnerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Runner is the Schema for the runners API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=runners,scope=Namespaced
type Runner struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RunnerSpec   `json:"spec,omitempty"`
	Status RunnerStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RunnerList contains a list of Runner
type RunnerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Runner `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Runner{}, &RunnerList{})
}
