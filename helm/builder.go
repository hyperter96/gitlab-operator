package helm

import (
	"fmt"

	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/releaseutil"
	"k8s.io/kubectl/pkg/scheme"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/settings"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/resource"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/support/charts"
)

// Builder provides an interface to build and render a Helm template.
type Builder interface {

	// Chart returns the Helm chart that will be rendered.
	Chart() *chart.Chart

	// Namespace returns namespace of the template.
	Namespace() string

	// SetNamespace sets namespace of the template. Changes will not take effect after rendering the
	// template.
	SetNamespace(namespace string)

	// ReleaseName returns release name of the template.
	ReleaseName() string

	// SetReleaseName sets release name of the template. Changes will not take effect after rendering
	// the template.
	SetReleaseName(releaseName string)

	// HooksDisabled returns true if hooks are disabled for the template.
	HooksDisabled() bool

	// DisableHooks disables hooks for the template. Changes will not take effect after rendering the
	// template.
	DisableHooks()

	// EnableHooks enables hooks for the template. Changes will not take effect after rendering the
	// template.
	EnableHooks()

	// Render renders the template with the provided values and parses the objects.
	Render(values resource.Values) (Template, error)
}

// NewBuilder creates a new builder interface for Helm template.
func NewBuilder(name, version string) (Builder, error) {
	envSettings := cli.New()

	actionConfig := new(action.Configuration)
	actionConfig, err := actionConfig, actionConfig.Init(
		envSettings.RESTClientGetter(), envSettings.Namespace(),
		memoryStorageDriver, noopLogger)

	if err != nil {
		return nil, err
	}

	client := action.NewInstall(actionConfig)
	client.DryRun = true
	client.Replace = true
	client.ClientOnly = true
	client.KubeVersion = settings.KubeVersion
	client.APIVersions = settings.GetKubeAPIVersions()

	chart := charts.GlobalCatalog().Query(
		charts.WithName(name), charts.WithVersion(version)).First()

	if chart == nil {
		return nil, errors.Errorf("chart %s-%s not found", name, version)
	}

	return &defaultBuilder{
		client:      client,
		chart:       chart,
		namespace:   envSettings.Namespace(),
		releaseName: defaultReleaseName,
	}, nil
}

const (
	defaultReleaseName  = "ephemeral"
	memoryStorageDriver = "memory"
)

var (
	noopLogger = func(_ string, _ ...interface{}) {}
)

type defaultBuilder struct {
	client       *action.Install
	chart        *chart.Chart
	namespace    string
	releaseName  string
	disableHooks bool
}

func (b *defaultBuilder) Chart() *chart.Chart {
	return b.chart
}

// Namespace returns namespace of the template.
func (b *defaultBuilder) Namespace() string {
	return b.namespace
}

// SetNamespace sets namespace of the template.
func (b *defaultBuilder) SetNamespace(namespace string) {
	b.namespace = namespace
}

// ReleaseName returns release name of the template.
func (b *defaultBuilder) ReleaseName() string {
	return b.releaseName
}

// SetReleaseName sets release name of the template.
func (b *defaultBuilder) SetReleaseName(releaseName string) {
	b.releaseName = releaseName
}

// HooksDisabled returns true if hooks are disabled in the template.
func (b *defaultBuilder) HooksDisabled() bool {
	return b.disableHooks
}

// DisableHooks disables hooks for the template.
func (b *defaultBuilder) DisableHooks() {
	b.disableHooks = true
}

// EnableHooks enables hooks for the template.
func (b *defaultBuilder) EnableHooks() {
	b.disableHooks = false
}

// Render renders the template with the provided values and parses the objects.
func (b *defaultBuilder) Render(values resource.Values) (Template, error) {
	b.client.DisableHooks = b.disableHooks
	b.client.Namespace = b.namespace
	b.client.ReleaseName = b.releaseName

	release, err := b.client.Run(b.chart, values)
	if err != nil {
		return nil, err
	}

	manifests := releaseutil.SplitManifests(release.Manifest)

	if !b.disableHooks {
		for index, hook := range release.Hooks {
			manifests[fmt.Sprintf("hook-%d", index)] =
				fmt.Sprintf("# Hook: %s\n%s\n", hook.Path, hook.Manifest)
		}
	}

	decode := scheme.Codecs.UniversalDeserializer().Decode

	template := newMutableTemplate(b.releaseName, b.namespace)

	for _, yaml := range manifests {
		obj, _, err := decode([]byte(yaml), nil, nil)

		if err != nil {
			template.warnings = append(template.warnings, err)
		} else {
			template.objects = append(template.objects, obj)
		}
	}

	return template, nil
}
