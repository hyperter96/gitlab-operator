/*


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

// RunnerSpec defines the desired state of Runner
type RunnerSpec struct {
	// GitlabResource represents a Gitlab custom resource. Should
	// only be used to reference Gitlab instance created by the operator
	Gitlab GitlabInstanceSpec `json:"gitlab,omitempty"`
	// RegistrationToken is name of secret with the
	// runner-registration-token key used to register the runner
	RegistrationToken string `json:"token,omitempty"`
	// Tags passes the runner tags
	Tags string `json:"tags,omitempty"`

	// Cache defines an S3 compatible object store
	Cache *RunnerCacheSpec `json:"cache,omitempty"`
}

// GitlabInstanceSpec defines the Gitlab custom
// resource in the kubernetes
type GitlabInstanceSpec struct {
	// Name of gitlab resource in kubernetes / openshift
	Name string `json:"name,omitempty"`
	// Gitlab or Continuous Integration URL
	URL string `json:"url,omitempty"`
}

// RunnerCacheSpec allows end user
// to define an S3 cache for the runner
type RunnerCacheSpec struct {
	// S3 cache server URL
	Server string `json:"server,omitempty"`

	// Region for the cache
	Region string `json:"region,omitempty"`

	// Credentials is the name of the secret containing the
	Credentials string `json:"credentials,omitempty"`

	// Insecure enables use of HTTP protocol
	Insecure bool `json:"insecure,omitempty"`

	// Path defines the Runner Cache path
	Path string `json:"path,omitempty"`

	// Bucket defines the s3 bucket name
	Bucket string `json:"bucket,omitempty"`
}

// RunnerStatus defines the observed state of Runner
type RunnerStatus struct {
	Phase        string `json:"phase,omitempty"`
	Registration string `json:"registration,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Runner is the Schema for the runners API
type Runner struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RunnerSpec   `json:"spec,omitempty"`
	Status RunnerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RunnerList contains a list of Runner
type RunnerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Runner `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Runner{}, &RunnerList{})
}
