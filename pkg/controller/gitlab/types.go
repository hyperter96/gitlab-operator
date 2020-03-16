package gitlab

import (
	corev1 "k8s.io/api/core/v1"
)

const (
	// GitlabEnterpriseImage represents the gitlab enterprise edition
	// Image to be deployed in our environment
	GitlabEnterpriseImage = "gitlab/gitlab-ee:12.8.6-ee.0"
	// GitlabCommunityImage represents the gitlab  Community
	// edition image to be deployed
	GitlabCommunityImage = "gitlab/gitlab-ce:12.8.6-ce.0"
	// GitlabRunnerImage represents the runner image
	GitlabRunnerImage = "gitlab/gitlab-runner:v12.8.0"
	// StrongPassword defines password length
	StrongPassword = 21
)

// Component represents an application / micro-service
// that makes part of a larger application
type Component struct {
	// Namespace of the component
	Namespace string
	// Defines the number of pods for the
	// component to be created
	Replicas int32
	// Labels for the component
	Labels map[string]string
	// InitContainers contains a list of containers	that may
	// need to run before the main application container starts up
	InitContainers []corev1.Container
	// Containers containers a list of containers that make up a pod
	Containers []corev1.Container
	// Contains a list of volumes used by the containers
	Volumes []corev1.Volume
	// Defines volume claims to be used by a statefulset
	VolumeClaimTemplates []corev1.PersistentVolumeClaim
}

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

type security interface {
	GenerateComponentPasswords()
	RunnerRegistrationToken() string
	GitlabRootPassword() string
	PostgresPassword() string
	RedisPassword() string
}

// PasswordOptions provides paramaters to be
// used when generating passwords
type PasswordOptions struct {
	// Length defines desired password length
	Length int
	// EnableSpecialCharacters adds special characters
	// to generated passwords
	EnableSpecialChars bool
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
