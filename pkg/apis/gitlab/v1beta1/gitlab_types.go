package v1beta1

import (
	acmev1alpha2 "github.com/jetstack/cert-manager/pkg/apis/acme/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GitlabSpec defines the desired state of Gitlab
type GitlabSpec struct {
	Replicas    int32               `json:"replicas"`
	Release     string              `json:"release,omitempty"`
	URL         string              `json:"url,omitempty"`
	TLS         string              `json:"tls,omitempty"`
	SMTP        SMTPConfiguration   `json:"email,omitempty"`
	Registry    RegistrySpec        `json:"registry,omitempty"`
	ObjectStore ObjectStoreSpec     `json:"objectStore,omitempty"`
	Redis       *RedisSpec          `json:"redis,omitempty"`
	Database    *DatabaseSpec       `json:"postgresql,omitempty"`
	CertIssuer  *ACMEOptions        `json:"acme,omitempty"`
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
	Disabled bool   `json:"disable,omitempty"`
	URL      string `json:"url,omitempty"`
	TLS      string `json:"tls,omitempty"`
}

// ObjectStoreSpec defines options for Gitlab registry
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
	// Password represents the password for SMTP user
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
	Capacity string `json:"capacity,omitempty"`
	// StorageClass from which volume should originate
	StorageClass string `json:"storageClass,omitempty"`
}

// ComponentVolumeSpec defines volumes for
// the different Gitlab peieces
type ComponentVolumeSpec struct {
	// Postgres database volume for PGDATA
	Postgres VolumeSpec `json:"database,omitempty"`
	// Redis key value store volume
	Redis VolumeSpec `json:"redis,omitempty"`
	// Gitlab registry volume
	Repositories VolumeSpec `json:"repositories,omitempty"`
}

// GitlabStatus defines the observed state of Gitlab
type GitlabStatus struct {
	// Phase represents status of the Gitlab resource
	Phase       string       `json:"phase,omitempty"`
	Release     string       `json:"release,omitempty"`
	Stage       string       `json:"stage,omitempty"`
	HealthCheck *HealthCheck `json:"health,omitempty"`
}

// HealthCheck represents the status
//  of services that make up Gitlab
type HealthCheck struct {
	Postgres  string `json:"database,omitempty"`
	Redis     string `json:"redis,omitempty"`
	Workhorse string `json:"workhorse,omitempty"`
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
