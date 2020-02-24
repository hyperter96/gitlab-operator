package gitlab

import (
	corev1 "k8s.io/api/core/v1"
)

// Component represents an application / micro-service
// that makes part of a larger application
type Component struct {
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
