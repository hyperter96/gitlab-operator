package helm

import (
	"fmt"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/releaseutil"
	"k8s.io/kubectl/pkg/scheme"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/settings"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/resource"
)

// Builder provides an interface to build and render a Helm template.
type Builder interface {

	// Chart return the source of the chart that will be rendered. This field is immutable and must
	// be specified when Builder is initialized.
	Chart() string

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
func NewBuilder(chart string) Builder {
	envSettings := cli.New()

	return &defaultBuilder{
		envSettings:   envSettings,
		chart:         chart,
		namespace:     envSettings.Namespace(),
		releaseName:   defaultReleaseName,
		storageDriver: memoryStorageDriver,
		debugLogger:   noopLogger,
	}
}

const (
	defaultReleaseName  = "ephemeral"
	memoryStorageDriver = "memory"
)

var (
	noopLogger = func(_ string, _ ...interface{}) {}
)

type defaultBuilder struct {
	envSettings   *cli.EnvSettings
	storageDriver string
	debugLogger   func(string, ...interface{})
	chart         string
	namespace     string
	releaseName   string
	disableHooks  bool
}

func (b *defaultBuilder) Chart() string {
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

func (b *defaultBuilder) newActionConfig() (*action.Configuration, error) {
	actionConfig := new(action.Configuration)
	return actionConfig, actionConfig.Init(
		b.envSettings.RESTClientGetter(), b.namespace, b.storageDriver, b.debugLogger)
}

// Render renders the template with the provided values and parses the objects.
func (b *defaultBuilder) Render(values resource.Values) (Template, error) {
	actionConfig, err := b.newActionConfig()
	if err != nil {
		return nil, err
	}

	client := action.NewInstall(actionConfig)
	client.DryRun = true
	client.Replace = true
	client.ClientOnly = true
	client.DisableHooks = b.disableHooks
	client.Namespace = b.namespace
	client.ReleaseName = b.releaseName

	client.KubeVersion = settings.KubeVersion
	client.APIVersions = settings.GetKubeAPIVersions()

	chartPath, err := client.ChartPathOptions.LocateChart(b.chart, b.envSettings)
	if err != nil {
		return nil, err
	}

	chartRequested, err := loader.Load(chartPath)
	if err != nil {
		return nil, err
	}

	release, err := client.Run(chartRequested, values)
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
