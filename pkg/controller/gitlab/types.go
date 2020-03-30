package gitlab

import (
	gitlabv1beta1 "gitlab.com/ochienged/gitlab-operator/pkg/apis/gitlab/v1beta1"
)

const (
	// StrongPassword defines password length
	StrongPassword = 21

	// DatabaseName defines name of GitLab database in PostgreSQL
	DatabaseName = "gitlabhq_production"

	// DatabaseUser defines name of user in PostgreSQL
	DatabaseUser = "gitlab"
)

// RedisConfig struct configures redis password
// and cluster configuration for large environments
type RedisConfig struct {
	// Defines the redis host to be used in the configuration
	Password string
	Cluster  bool
}

// ComponentPasswords struct has passwords for the different
// gitlab components
type ComponentPasswords struct {
	redis                   string
	postgres                string
	gitlabRootPassword      string
	runnerRegistrationToken string
}

// OmnibusOptions defines options for
// configuring the gitlab pod
type OmnibusOptions struct {
	// Enable gitlab registry
	RegistryEnabled bool
	// RegistryExternalURL defines gitlab
	// registry external URL
	RegistryExternalURL string
	// MontiringNetworks contains a list of networks
	// That should be allowed to access gitlab metrics,
	// liveness probe and readiness probe endpoints
	MonitoringWhitelist string
	// SMTP server options
	SMTP gitlabv1beta1.SMTPConfiguration
}

// ReadinessStatus shows status of Gitlab services
type ReadinessStatus struct {
	// Returns status of Gitlab rails app
	WorkhorseStatus string `json:"status,omitempty"`
	// RedisStatus reports status of redis
	RedisStatus []ServiceStatus `json:"redis_check,omitempty"`
	// DatabaseStatus reports status of postgres
	DatabaseStatus []ServiceStatus `json:"db_check,omitempty"`
}

// ServiceStatus shows status of a Gitlab
// dependent service .e.g. Postgres, Redis, Gitaly
type ServiceStatus struct {
	Status string `json:"status,omitempty"`
}

// GitalyOptions contains service
// names for Redis and Unicorn
type GitalyOptions struct {
	// Name of redis service
	RedisMaster string

	// Name of Unicorn service
	Unicorn string
}

// UnicornOptions passes options
// to unicorn templates
type UnicornOptions struct {
	Namespace   string
	GitlabURL   string
	PostgreSQL  string
	Registry    string
	RegistryURL string
	Minio       string
	MinioURL    string
	Gitaly      string
	RedisMaster string
	EmailFrom   string
	ReplyTo     string
}

// WorkhorseOptions has
// options for workhorse
type WorkhorseOptions struct {
	RedisMaster string
}

// ShellOptions passes template
// options for gitlab shell
type ShellOptions struct {
	Unicorn     string
	RedisMaster string
}

// SidekiqOptions defines parameters
// for sidekiq configmap
type SidekiqOptions struct {
	RedisMaster    string
	PostgreSQL     string
	GitlabDomain   string // ExternalURL no protocol. e.g: gitlab.example.com
	EnableRegistry bool
	Registry       string
	RegistryDomain string
	Gitaly         string
	Namespace      string
	EmailFrom      string
	ReplyTo        string
	MinioDomain    string // hostname e.g. minio.example.com
	Minio          string // Minio service
}

// ExporterOptions defines parameters
// for metrics exporter configmap
type ExporterOptions struct {
	RedisMaster string
	Postgres    string
}

// RegistryOptions defines parameters
// for registry configmap
type RegistryOptions struct {
	GitlabDomain string
	Minio        string
}

// RailsOptions defines parameters
// for rails secret
type RailsOptions struct {
	SecretKey     string
	DatabaseKey   string
	OTPKey        string
	RSAPrivateKey []string
}

// TaskRunnerOptions defines options
// for Task Runner configurations
type TaskRunnerOptions struct {
	RedisMaster string
	Namespace   string
	GitlabURL   string
	Minio       string
	MinioURL    string
	Registry    string
	RegistryURL string
	Gitaly      string
	MailFrom    string
	ReplyTo     string
	PostgreSQL  string
}

// ConfigOptions has options for
// Redis and Postgres configs
type ConfigOptions struct {
	RedisMaster string
	Postgres    string
}

// MigrationOptions provides options
// required by the migrations job
type MigrationOptions struct {
	Namespace   string
	RedisMaster string
	PostgreSQL  string
	Gitaly      string
}
