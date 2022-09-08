package gitlab

import (
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/resource"
)

// Adapter is a purpose-built wrapper for GitLab resources. It provides a
// convenient interface to interact with GitLab resources.
//
// Use internal `NewAdapter` functions to create a new wrapper for a specific
// GitLab resource version, for example `internal.NewV1Beta1Adapter`.
type Adapter interface {
	Operation
	Features
	Status
	resource.CustomResourceWrapper
	resource.ValueProvider
	resource.ChartConsumer
}
