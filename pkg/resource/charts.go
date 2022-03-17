package resource

import (
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/support/charts"
)

// ChartConsumer offers a mechanism to associate specific Helm Charts to an
// individual Custom Resource that consumes them to reconcile its state.
//
// The provider is a Custom Resource wrapper that can identify the required
// Helm Charts based on the specification of the underlying resource. Each
// provider can have different capabilities which impact their ability to
// locate and load the required Helm Charts.
type ChartConsumer interface {
	// ReleaseName returns the name of the Helm release of this resource. This
	// name is shared between all Helm Charts.
	ReleaseName() string

	// Charts returns all the Helm Charts that the controller uses to reconcile
	// the state of the associated resource.
	//
	// The provider infers the required Helm Charts from the underlying Custom
	// Resource and tries to locate and load them.
	//
	// It returns and error if it fails to identify, locate, or load any of the
	// required Helm Charts.
	Charts() (charts.Catalog, error)
}
