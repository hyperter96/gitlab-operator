package v1beta1

import (
	_ "embed"

	"context"
	"html/template"
	"strings"

	"github.com/mitchellh/copystructure"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/chartutil"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/settings"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/support"
)

/* ValueProvider */

func (w *Adapter) Values() support.Values {
	return w.values
}

func (w *Adapter) Hash() string {
	return support.SimpleObjectHash(w.source)
}

/* Helpers */

func (w *Adapter) populate(ctx context.Context) error {
	return support.ChainedOperation{
		w.applyUserDefinedValues,
		w.applyOperatorDefaultValues,
		w.fillMissingValues,
		w.applyChartDefaultValues, // it uses coalesce (set value if not present)
	}.Run(ctx)
}

func (w *Adapter) applyUserDefinedValues(_ context.Context) error {
	val, err := copystructure.Copy(w.source.Spec.Chart.Values.Object)
	if err != nil {
		return errors.Wrap(err, "failed to clone user-defined values")
	}

	w.values = val.(map[string]interface{})
	if w.values == nil {
		w.values = support.Values{}
	}

	return nil
}

func (w *Adapter) applyOperatorDefaultValues(_ context.Context) error {
	var parameters = map[string]interface{}{
		"ReleaseName":    w.ReleaseName(),
		"UseCertManager": w.WantsFeature(ConfigureCertManager),
		"ExternalIP":     w.values.GetString("global.hosts.externalIP"),
		"Settings":       appSettings,
	}

	var defaultValues *strings.Builder = &strings.Builder{}
	if err := defaultValuesTemplate.Execute(defaultValues, parameters); err != nil {
		return errors.Wrap(err, "failed to render operator default values")
	}

	if err := w.values.AddFromYAML(defaultValues.String()); err != nil {
		return errors.Wrap(err, "can not merge operator default values")
	}

	return nil
}

func (w *Adapter) applyChartDefaultValues(_ context.Context) error {
	charts, err := w.Charts()
	if err != nil {
		return err
	}

	for _, chart := range charts {
		val, err := chartutil.CoalesceValues(chart, w.values)
		if err != nil {
			return errors.Wrapf(err, "failed to coalesce chart values: %s", chart.Name())
		}

		w.values = val.AsMap()
	}

	return nil
}

func (w *Adapter) fillMissingValues(_ context.Context) error {
	if email := w.values.GetString("certmanager-issuer.email"); email == "" {
		_ = w.values.SetValue("certmanager-issuer.email", defaultCertManagerIssuerEmail)
	}

	return nil
}

//go:embed values.tpl
var defaultValuesSource string
var defaultValuesTemplate *template.Template
var appSettings map[string]string

func init() {
	settings.Load()

	defaultValuesTemplate = template.Must(template.New("defaultValues").Parse(defaultValuesSource))
	appSettings = map[string]string{
		"AppNonRootServiceAccount": settings.AppNonRootServiceAccount,
		"AppAnyUIDServiceAccount":  settings.AppAnyUIDServiceAccount,
		"ManagerServiceAccount":    settings.ManagerServiceAccount,
		"NginxServiceAccount":      settings.NGINXServiceAccount,
	}
}
