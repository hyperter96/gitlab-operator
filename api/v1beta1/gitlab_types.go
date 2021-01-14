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
	acmev1alpha2 "github.com/jetstack/cert-manager/pkg/apis/acme/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GitLabSpec defines the desired state of GitLab
type GitLabSpec struct {
	// The GitLab version to deploy
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Release",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	Release string `json:"release,omitempty"`

	// The fully qualified domain name used to access the GitLab instance.
	// For example: gitlab.example.com
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="GitLab URL",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	URL string `json:"url,omitempty"`
	// Name of tls secret used to secure the GitLab instance URL
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="TLS Certificate",xDescriptors="urn:alm:descriptor:com.tectonic.ui:selector:core:v1:Secret"
	TLS string `json:"tls,omitempty"`
	// If specified, SMTP provides the details of the email server
	// used by GitLab to send outgoing email
	SMTP SMTPConfiguration `json:"smtp,omitempty"`
	// Options used to setup the GitLab Registry
	Registry RegistrySpec `json:"registry,omitempty"`
	// The parameters for the object storage used to store GitLab artifacts
	ObjectStore ObjectStoreSpec `json:"objectStore,omitempty"`
	// If specified, the Redis options override the default behavior of the
	// Redis key-value store deployed by the operator
	Redis *RedisSpec `json:"redis,omitempty"`
	// If specified, overrides the default behavior of the Postgresql
	// database deployed by the operator
	Database *DatabaseSpec `json:"postgres,omitempty"`
	// If specified, the options used by Cert-Manager to generate certificates.
	// More info: https://cert-manager.io/docs/configuration/acme/
	CertIssuer *ACMEOptions `json:"acme,omitempty"`
	// Volume for Gitaly statefulset
	Volume VolumeSpec `json:"volume,omitempty"`
	// If specified, defines the parameters used when autoscaling GitLab resources
	AutoScaling *AutoScalingSpec `json:"autoscaling,omitempty"`
}

// RedisSpec defines Redis options
type RedisSpec struct {
	Replicas int32      `json:"replicas,omitempty"`
	Volume   VolumeSpec `json:"volume,omitempty"`
}

// DatabaseSpec defines database options
type DatabaseSpec struct {
	Replicas int32      `json:"replicas,omitempty"`
	Volume   VolumeSpec `json:"volume,omitempty"`
}

// RegistrySpec defines options for GitLab registry
type RegistrySpec struct {
	Disabled bool   `json:"disable,omitempty"`
	URL      string `json:"url,omitempty"`
	TLS      string `json:"tls,omitempty"`
}

// ObjectStoreSpec defines options for GitLab registry
type ObjectStoreSpec struct {
	// Development will result in a minio deployment being
	// created for testing /development purposes
	Development bool `json:"development,omitempty"`
	// URL provides a domain / DNS name that can be used
	// to reach the minio deployment
	URL string `json:"url,omitempty"`
	// Credentials contains the name of the secret that contains
	// the `accesskey` and `secretkey` values required to access
	// an existing minio instance. Should be an even number equal
	// to or larger than four
	Credentials string `json:"credentials,omitempty"`
	// StorageClass defines the storage class for the persistent
	// volume that will hold the s3 objects
	// +optional
	StorageClass string `json:"storageClass,omitempty"`
}

// SMTPConfiguration defines options for Gitlab registry
// More on configuration options available below:
// https://docs.gitlab.com/omnibus/settings/smtp.html
type SMTPConfiguration struct {
	// Host is the SMTP host
	Host string `json:"host,omitempty"`
	// Port represents SMTP port
	Port int32 `json:"port,omitempty"`
	// Domain represents the email domain
	Domain string `json:"domain,omitempty"`
	// Username represents the SMTP username for sending email
	Username string `json:"username,omitempty"`
	// Password contains name of secret containing
	// the password for SMTP user
	Password string `json:"password,omitempty"`
	// Authentication represents authentication mechanism
	// Options include: login, plain, cram_md5
	Authentication string `json:"authentication,omitempty"`
	// EnableSSL enables/disables SSL/TLS
	EnableSSL bool `json:"enableSSL,omitempty"`
	// EnableStartTLS enables starttls
	EnableStartTLS bool `json:"enableStartTLS,omitempty"`
	// OpenSSLVerifyMode sets how OpenSSL checks the
	// certificate whenever TLS is used
	// OpenSSLVerifyMode can be: 'none', 'peer'
	OpenSSLVerifyMode string `json:"opensslVerifyMode,omitempty"`
	// EmailFrom represents the from address of outgoing email
	EmailFrom string `json:"from,omitempty"`
	// ReplyTO specifies a reply to email address
	ReplyTo string `json:"replyTo,omitempty"`
	// DisplayName represents the name of the email
	DisplayName string `json:"displayName,omitempty"`
}

// AutoScalingSpec are the parameters to configure autoscaling
type AutoScalingSpec struct {
	// Minimum number of replicas to scale to
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Minimum Replicas",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	MinReplicas *int32 `json:"minReplicas,omitempty"`
	// Maximum number of replicas to scale to
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Maxiumum Replicas",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	MaxReplicas int32 `json:"maxReplicas,omitempty"`
	// Percentage CPU of the requested CPU resources at which autoscaling triggers
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="CPU Percentage Threshold",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	TargetCPU *int32 `json:"targetCPU,omitempty"`
}

// ACMEOptions defines the values for the
// ACME service that will provide certificates
type ACMEOptions struct {
	// Email is the email for this account
	// +optional
	Email string `json:"email,omitempty"`

	// Server is the ACME server URL
	// Default to letsencrypt production URL
	//+optional
	Server string `json:"server,omitempty"`

	// If true, skip verifying the ACME server TLS certificate
	// +optional
	SkipTLSVerify bool `json:"skipTLSVerify,omitempty"`

	// ExternalAccountBinding is a reference to a CA external account of the ACME
	// server.
	// +optional
	ExternalAccountBinding *acmev1alpha2.ACMEExternalAccountBinding `json:"externalAccountBinding,omitempty"`

	// Solvers is a list of challenge solvers that will be used to solve
	// ACME challenges for the matching domains.
	// +optional
	Solvers []acmev1alpha2.ACMEChallengeSolver `json:"solvers,omitempty"`
}

// VolumeSpec defines volume specifications
type VolumeSpec struct {
	// Capacity of the volume
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Storage capacity",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	Capacity string `json:"capacity,omitempty"`
	// StorageClass from which volume should originate
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Storage class",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	StorageClass string `json:"storageClass,omitempty"`
}

// HealthCheck represents the status
//  of services that make up Gitlab
type HealthCheck struct {
	Postgres  string `json:"database,omitempty"`
	Redis     string `json:"redis,omitempty"`
	Workhorse string `json:"workhorse,omitempty"`
}

// GitLabStatus defines the observed state of GitLab
type GitLabStatus struct {
	Phase       string       `json:"phase,omitempty"`
	Release     string       `json:"release,omitempty"`
	Stage       string       `json:"stage,omitempty"`
	HealthCheck *HealthCheck `json:"health,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=gl
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="STATUS",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="REDIS",type=string,JSONPath=`.status.health.redis`
// +kubebuilder:printcolumn:name="DATABASE",type=string,JSONPath=`.status.health.database`
// +kubebuilder:printcolumn:name="CONSOLE",type=string,JSONPath=`.status.health.workhorse`
// +operator-sdk:csv:customresourcedefinitions:displayName="GitLab"
// +operator-sdk:csv:customresourcedefinitions:resources={{ConfigMap,v1,""},{Secret,v1,""},{Service,v1,""},{Pod,v1,""},{Deployment,v1,""},{StatefulSet,v1,""},{PersistentVolumeClaim,v1,""},{Runner,v1beta1,""},{GLBackup,v1beta1,""}}

// GitLab is a complete DevOps platform, delivered in a single application
type GitLab struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of a GitLab instance
	Spec GitLabSpec `json:"spec,omitempty"`
	// Most recently observed status of the GitLab instance.
	// It is read-only to the user
	Status GitLabStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GitLabList contains a list of GitLab
type GitLabList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GitLab `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GitLab{}, &GitLabList{})
}
