package gitlab

import (
	corev1 "k8s.io/api/core/v1"
)

const (
	// GitlabEnterpriseImage represents the gitlab enterprise edition
	// Image to be deployed in our environment
	GitlabEnterpriseImage = "gitlab/gitlab-ee:12.8.0-ee.0"
	// GitlabCommunityImage represents the gitlab  Community
	// edition image to be deployed
	GitlabCommunityImage = "gitlab/gitlab-ce:12.6.7-ce.0"
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
	GeneratePasswords()
	RunnerRegistrationToken() string
	GitlabRootPassword() string
	PostgresPassword() string
	RedisPassword() string
}
