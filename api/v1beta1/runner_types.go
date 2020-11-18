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
	// gitlab specifies the GitLab instance the GitLab Runner
	// will register against
	GitLab GitLabInstanceSpec `json:"gitlab"`

	//Name of secret containing the 'runner-registration-token' key used to register the runner
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Registration Token",xDescriptors="urn:alm:descriptor:com.tectonic.ui:selector:core:v1:Secret"
	RegistrationToken string `json:"token,omitempty"`

	// List of comma separated tags to be applied to the runner
	// More info: https://docs.gitlab.com/ee/ci/runners/#use-tags-to-limit-the-number-of-jobs-using-the-runner
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Tags",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	Tags string `json:"tags,omitempty"`

	// Option to limit the number of jobs globally that can run concurrently.
	// The operator sets this to 10, if not specified
	// +operator-sdk:csv:customresourcedefinitions:type=status,displayName="Concurrent",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	Concurrent *int32 `json:"concurrent,omitempty"`

	// Option to define the number of seconds between checks for new jobs.
	// This is set to a default of 30s by operator if not set
	// +operator-sdk:csv:customresourcedefinitions:type=status,displayName="Check Interval",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	CheckInterval *int32 `json:"interval,omitempty"`

	// If specified, overrides the default URL used to clone or fetch the Git ref
	CloneURL string `json:"cloneURL,omitempty"`

	// If specified, overrides the default GitLab Runner helper image
	HelperImage string `json:"helperImage,omitempty"`

	// The name of the default image to use to run
	// build jobs, when none is specified
	BuildImage string `json:"buildImage,omitempty"`

	// Type of cache used for Runner artifacts
	// Options are: gcs, s3, azure
	// +kubebuilder:validations:Enum=s3;gcs;azure
	CacheType string `json:"cacheType,omitempty"`

	// Path defines the Runner Cache path
	CachePath string `json:"cachePath,omitempty"`

	// Name of tls secret containing the custom certificate
	// authority (CA) certificates
	CertificateAuthority string `json:"ca,omitempty"`

	// Enable sharing of cache between Runners
	CacheShared bool `json:"cacheShared,omitempty"`

	// options used to setup S3
	// object store as GitLab Runner Cache
	S3 *CacheS3Config `json:"s3,omitempty"`
	// options used to setup GCS (Google
	// Container Storage) as GitLab Runner Cache
	GCS *CacheGCSConfig `json:"gcs,omitempty"`
	// options used to setup Azure blob
	// storage as GitLab Runner Cache
	Azure *CacheAzureConfig `json:"azure,omitempty"`
	// allow user to override service account
	// used by GitLab Runner
	ServiceAccount string `json:"serviceaccount,omitempty"`
}

// GitLabInstanceSpec defines the Gitlab custom
// resource in the kubernetes
type GitLabInstanceSpec struct {
	// Name of GitLab instance created by the operator
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Instance Name",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	Name string `json:"name,omitempty"`
	// The fully qualified domain name of the address used to access the GitLab instance.
	// For example, gitlab.example.com
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Instance URL",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	URL string `json:"url,omitempty"`
}

// CacheS3Config defines options for an S3 compatible cache
type CacheS3Config struct {
	Server string `json:"server,omitempty"`
	// Name of the secret containing the
	// 'accesskey' and 'secretkey' used to access the object storage
	Credentials string `json:"credentials,omitempty"`
	// Name of the bucket in which the cache will be stored
	BucketName string `json:"bucket,omitempty"`
	// Name of the S3 region in use
	BucketLocation string `json:"location,omitempty"`
	// Use insecure connections or HTTP
	Insecure bool `json:"insecure,omitempty"`
}

// CacheGCSConfig defines options for GCS object store
type CacheGCSConfig struct {
	// contains the GCS 'access-id' and 'private-key'
	Credentials string `json:"credentials,omitempty"`
	// Takes GCS credentials file, 'keys.json'
	CredentialsFile string `json:"credentialsFile,omitempty"`
	// Name of the bucket in which the cache will be stored
	BucketName string `json:"bucket,omitempty"`
}

// CacheAzureConfig defines options for Azure object store
type CacheAzureConfig struct {
	// Credentials secret contains 'accountName' and 'privateKey'
	// used to authenticate against Azure blob storage
	Credentials string `json:"credentials,omitempty"`
	// Name of the Azure container in which the cache will be stored
	ContainerName string `json:"container,omitempty"`
	// The domain name of the Azure blob storage
	// e.g. blob.core.windows.net
	StorageDomain string `json:"storageDomain,omitempty"`
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

// Runner is the open source project used to run your jobs and send the results back to GitLab
type Runner struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of a GitLab Runner instance
	Spec RunnerSpec `json:"spec,omitempty"`
	// Most recently observed status of the GitLab Runner.
	// It is read-only to the user
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
