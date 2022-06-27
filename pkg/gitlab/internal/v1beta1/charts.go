package v1beta1

import (
	"context"
	"strings"

	semver "github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/gitlab"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/gitlab/component"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/support/charts"
)

/* ChartConsumer */

func (w *Adapter) ReleaseName() string {
	return w.source.Name
}

func (w *Adapter) Charts() (charts.Catalog, error) {
	return w.charts, nil
}

/* Helpers */

func (w *Adapter) prepareCharts(_ context.Context) error {
	name := component.GitLab.Name()
	version := w.source.Spec.Chart.Version

	if _, err := semver.NewVersion(version); err != nil {
		return errors.Wrapf(err, "invalid version format %s", version)
	}

	result := gitlab.GetChartCatalog().Query(
		charts.WithName(name), charts.WithVersion(version))

	if result.Empty() {
		return errors.Errorf("%s chart version %s not found; use one of %s",
			name, version, strings.Join(gitlab.GetChartCatalog().Versions(name), ", "))
	}

	w.charts = result

	return nil
}
