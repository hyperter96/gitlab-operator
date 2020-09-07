package helm

import (
	"fmt"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/releaseutil"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
)

// Template represents a Helm template object editor.
type Template struct {
	envSettings   *cli.EnvSettings
	storageDriver string
	chartName     string
	namespace     string
	releaseName   string
	disableHooks  bool
	objects       []*runtime.Object
	logger        func(string, ...interface{})
}

// ObjectSelector represents a boolean expression for selecting objects
type ObjectSelector = func(runtime.Object) bool

// ObjectEditor represents a method for editing objects
type ObjectEditor = func(*runtime.Object) error

// DefaultReleaseName is the default name that is used in the Helm release.
const DefaultReleaseName = "ephemeral"
const memoryStorageDriver = "memory"

var nopLogger = func(_ string, _ ...interface{}) {}

// NewTemplate creates a new Helm template object editor.
func NewTemplate(chartName string) *Template {
	envSettings := cli.New()
	return &Template{
		envSettings:   envSettings,
		chartName:     chartName,
		namespace:     envSettings.Namespace(),
		releaseName:   DefaultReleaseName,
		objects:       []*runtime.Object{},
		storageDriver: memoryStorageDriver,
		logger:        nopLogger,
	}
}

// ChartName returns chart name of the template.
func (t *Template) ChartName() string {
	return t.chartName
}

// Namespace returns namespace of the template.
func (t *Template) Namespace() string {
	return t.namespace
}

// SetNamespace sets namespace of the template.
func (t *Template) SetNamespace(namespace string) {
	t.namespace = namespace
}

// ReleaseName returns release name of the template.
func (t *Template) ReleaseName() string {
	return t.releaseName
}

// SetReleaseName sets release name of the template.
func (t *Template) SetReleaseName(releaseName string) {
	t.releaseName = releaseName
}

// HooksDisabled returns true if hooks are disabled in the template.
func (t *Template) HooksDisabled() bool {
	return t.disableHooks
}

// DisableHooks disables hooks for the template.
func (t *Template) DisableHooks() {
	t.disableHooks = true
}

// EnableHooks enables hooks for the template.
func (t *Template) EnableHooks() {
	t.disableHooks = false
}

// Objects returns list of available objects.
func (t *Template) Objects() []*runtime.Object {
	return t.objects
}

func (t *Template) newActionConfig() (*action.Configuration, error) {
	actionConfig := new(action.Configuration)
	return actionConfig, actionConfig.Init(
		t.envSettings.RESTClientGetter(), t.namespace, t.storageDriver, t.logger)
}

// Load renders the template with the provided values and parses the objects.
func (t *Template) Load(values *Values) ([]error, error) {
	actionConfig, err := t.newActionConfig()
	if err != nil {
		return nil, err
	}

	client := action.NewInstall(actionConfig)
	client.DryRun = true
	client.Replace = true
	client.ClientOnly = true
	client.DisableHooks = t.disableHooks
	client.Namespace = t.namespace
	client.ReleaseName = t.releaseName

	chartPath, err := client.ChartPathOptions.LocateChart(t.chartName, t.envSettings)
	if err != nil {
		return nil, err
	}

	chartRequested, err := loader.Load(chartPath)
	if err != nil {
		return nil, err
	}

	release, err := client.Run(chartRequested, values.AsMap())
	if err != nil {
		return nil, err

	}

	manifests := releaseutil.SplitManifests(release.Manifest)

	if !t.disableHooks {
		for index, hook := range release.Hooks {
			manifests[fmt.Sprintf("hook-%d", index)] =
				fmt.Sprintf("# Hook: %s\n%s\n", hook.Path, hook.Manifest)
		}
	}

	decode := scheme.Codecs.UniversalDeserializer().Decode
	warnings := []error{}

	for _, yaml := range manifests {
		obj, _, err := decode([]byte(yaml), nil, nil)

		if err != nil {
			warnings = append(warnings, err)
		} else {
			t.objects = append(t.objects, &obj)
		}
	}

	return warnings, nil
}

// GetObjects returns all objects that match the selector.
func (t *Template) GetObjects(selector ObjectSelector) ([]*runtime.Object, error) {
	result := []*runtime.Object{}
	for i := 0; i < len(t.objects); i++ {
		if selector(*t.objects[i]) {
			result = append(result, t.objects[i])
		}
	}
	return result, nil
}

// AddObject adds a new object to the template.
func (t *Template) AddObject(object runtime.Object) error {
	t.objects = append(t.objects, &object)

	return nil
}

// DeleteObjects deletes all objects that match the selector.
func (t *Template) DeleteObjects(selector ObjectSelector) (int, error) {
	count := 0
	for i := 0; i < len(t.objects); i++ {
		if selector(*t.objects[i]) {
			t.objects = append(t.objects[:i], t.objects[i+1:]...)
			count++
			i--
		}
	}
	return count, nil
}

// ReplaceObject replaces the first object that matches the selector with the new object.
func (t *Template) ReplaceObject(selector ObjectSelector, object runtime.Object) (*runtime.Object, error) {
	for i := 0; i < len(t.objects); i++ {
		if selector(*t.objects[i]) {
			old := t.objects[i]
			t.objects[i] = &object
			return old, nil
		}
	}
	return nil, nil
}

// EditObjects edits all objects that the editor can handle.
func (t *Template) EditObjects(editor ObjectEditor) (int, error) {
	count := 0
	for i := 0; i < len(t.objects); i++ {
		err := editor(t.objects[i])
		if err != nil {
			if IsTypeMistmatchError(err) {
				continue
			}
			return count, err
		}
		count++
	}
	return count, nil
}
