package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GitlabSpec defines the desired state of Gitlab
type GitlabSpec struct {
	Replicas int32               `json:"replicas"`
	URL      string              `json:"url,omitempty"`
	TLS      string              `json:"tlsSecret,omitempty"`
	SMTP     SMTPConfiguration   `json:"email,omitempty"`
	Registry RegistrySpec        `json:"registry,omitempty"`
	Minio    *MinioSpec          `json:"minio,omitempty"`
	Redis    *RedisSpec          `json:"redis,omitempty"`
	Database *DatabaseSpec       `json:"postgresql,omitempty"`
	Volumes  ComponentVolumeSpec `json:"volumes,omitempty"`
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
	TLS      string `json:"tlsSecret,omitempty"`
}

// MinioSpec defines options for Gitlab registry
type MinioSpec struct {
	Disabled bool `json:"disable,omitempty"`
	// URL provides a domain / DNS name that can be used
	// to reach the minio deployment
	URL string `json:"url,omitempty"`
	// Replicas dictates the number of minio nodes to deploy
	Replicas int32 `json:"replicas,omitempty"`
	// TLS is the name of the secret containing the tls certificate
	// used to secure the minio endpoint
	TLS string `json:"tlsSecret,omitempty"`
	// Credentials contains the name of the secret that contains
	// the `accesskey` and `secretkey` values required to access
	// an existing minio instance. Should be an even number equal
	// to or larger than four
	Credentials string `json:"credentials,omitempty"`
	// Capacity defines the storage volume capacity used by each
	// minio node. E.g. if the size is set to 10Gi and you have 4
	// replicas, a total of 40Gi will be required by minio
	Capacity string `json:"capacity,omitempty"`
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
	Phase       string       `json:"phase,omitempty"`
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
