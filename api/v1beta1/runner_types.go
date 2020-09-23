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

	//Name of secret containing the runner-registration-token key used to register the runner
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Registration Token",xDescriptors="urn:alm:descriptor:com.tectonic.ui:selector:core:v1:Secret"
	RegistrationToken string `json:"token,omitempty"`

	// List of comma separated tags to be applied to the runner
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Tags",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	Tags string `json:"tags,omitempty"`

	// Concurrent limits the number of jobs globally that can run concurrently
	// +operator-sdk:csv:customresourcedefinitions:type=status,displayName="Concurrent",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	Concurrent *int32 `json:"concurrent,omitempty"`

	// CheckInterval defines the number of seconds between checks for new jobs
	// +operator-sdk:csv:customresourcedefinitions:type=status,displayName="Check Interval",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	CheckInterval *int32 `json:"interval,omitempty"`

	// Cache defines an S3 compatible object store
	Cache *RunnerCacheSpec `json:"cache,omitempty"`
}

// GitlabInstanceSpec defines the Gitlab custom
// resource in the kubernetes
type GitlabInstanceSpec struct {
	// Name of GitLab instance created by the operator
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Instance Name",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	Name string `json:"name,omitempty"`
	// URL of GitLab instance
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Instance URL",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
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
	// Reports status of the GitLab Runner instance
	// +operator-sdk:csv:customresourcedefinitions:type=status,displayName="Phase",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	Phase string `json:"phase,omitempty"`

	// Reports status of GitLab Runner registration
	// +operator-sdk:csv:customresourcedefinitions:type=status,displayName="Registration",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	Registration string `json:"registration,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +operator-sdk:csv:customresourcedefinitions:displayName="GitLab Runner"
// +operator-sdk:csv:customresourcedefinitions:resources={{ConfigMap,v1,""},{Secret,v1,""},{Service,v1,""},{Replicasets,v1,""},{Pod,v1,""},{Deployment,v1,""},{PersistentVolumeClaim,v1,""}}

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
