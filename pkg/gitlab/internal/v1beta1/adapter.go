package v1beta1

import (
	"context"

	api "gitlab.com/gitlab-org/cloud-native/gitlab-operator/api/v1beta1"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/support"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/support/charts"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/support/objects"
)

// GitLabAdapter for v1beta1.
//
// See GitLabAdapter documentation.
type Adapter struct {
	source *api.GitLab
	values support.Values
	charts charts.Catalog

	// NOTE: This is a temporary solution to migrate the existing Helm facility
	//       to the new framework. It will be removed once the migration is
	//       completed. Do not use it for any other purpose.
	targetManagedObjects objects.Collection
}

func NewAdapter(ctx context.Context, src *api.GitLab) (*Adapter, error) {
	adapter := &Adapter{
		source: src,
		values: support.Values{},

		targetManagedObjects: objects.Collection{},
	}

	return adapter, support.ChainedOperation{
		adapter.prepare,
		adapter.populate,
		adapter.validate,
	}.Run(ctx)
}

/* Helpers */

func (w *Adapter) prepare(ctx context.Context) error {
	return support.ChainedOperation{
		w.prepareCharts,
	}.Run(ctx)
}

func (w *Adapter) validate(_ context.Context) error {
	return nil
}
