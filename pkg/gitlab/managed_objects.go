package gitlab

import (
	"context"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/support/objects"
)

// ManagedObjects provides an access mechanism to the managed resources that the
// a Custom Resource resource requires to operate.
//
// The resource may not exist in the cluster or their state may differ from
// what is known to the controller.
//
// To retrieve the list of the current resources that are owned by the
// underlying GitLab resource, use Current function.
//
// To retrieve the list of resources that the controller wants to create or
// update in order to achieve the desired state of the underlying GitLab
// resource.
type ManagedObjects interface {
	// CurrentObjects returns the list of managed resources that are currently
	// owned by the underlying GitLab resource.
	//
	// For a new installation this list can be empty.
	CurrentObjects(ctx context.Context) (objects.Collection, error)

	// TargetObjects returns the list of managed resources that the controller
	// expects to be available to reconcile the state of the the underlying
	// GitLab resource.
	TargetObjects() objects.Collection
}
