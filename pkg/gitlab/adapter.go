package gitlab

import (
	"k8s.io/apimachinery/pkg/runtime"

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
	ManagedObjects
	resource.CustomResourceWrapper
	resource.ValueProvider
	resource.ChartConsumer

	// PopulateManagedObjects appends the list of objects to the list of target
	// managed resources.
	//
	// NOTE: This is a helper method to migrate the existing Helm facility to
	//       the new framework. It will be removed once the migration is
	//       completed. Do not use it for any other purpose.
	PopulateManagedObjects(...runtime.Object) error
}
