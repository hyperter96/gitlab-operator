package helm

import (
	"os"
	"sync"

	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	instanceLock = &sync.Mutex{}
	theTemplate  Template

	// Logger is the main logger of this module.
	Logger = ctrl.Log.WithName("helm")
)

// GetTemplate ensures that only one instance of Helm template exists and it is loaded only once.
// This method uses the following environment variables:
//
//    HELM_CHART			The path to the Chart directory. Default: "gitlab"
//    HELM_VALUES			The path to the values YAML file. Default: "values.yaml"
//    HELM_NAMESPACE		The release namespace. Default: "default"
//    HELM_RELEASE_NAME		The release name. Default: "ephemeral"
//    HELM_DISABLE_HOOKS	Disable Helm hooks. Default: "false"
func GetTemplate() (Template, error) {
	instanceLock.Lock()
	defer instanceLock.Unlock()

	if theTemplate == nil {
		chartName := os.Getenv("HELM_CHART")
		if chartName == "" {
			chartName = "gitlab"
		}

		valuesYaml := os.Getenv("HELM_VALUES")
		if valuesYaml == "" {
			valuesYaml = "values.yaml"
		}

		namespace := os.Getenv("HELM_NAMESPACE")
		if namespace == "" {
			namespace = "default"
		}

		releaseName := os.Getenv("HELM_RELEASE_NAME")
		if releaseName == "" {
			releaseName = defaultReleaseName
		}

		disableHooks := os.Getenv("HELM_DISABLE_HOOKS")
		if disableHooks == "" {
			disableHooks = "false"
		}

		Logger.Info("Creating a new Helm template",
			"chart", chartName,
			"namespace", namespace,
			"releaseName", releaseName,
			"disableHooks", disableHooks)

		builder := NewBuilder(chartName)
		builder.SetNamespace(namespace)
		builder.SetReleaseName(releaseName)
		if disableHooks == "true" {
			builder.DisableHooks()
		}

		Logger.Info("Using values to render Helm template",
			"source", valuesYaml)

		values := EmptyValues()
		if err := values.AddFromFile(valuesYaml); err != nil {
			return nil, err
		}

		var err error
		theTemplate, err = builder.Render(values)

		if err != nil {
			return nil, err
		}

		Logger.Info("Helm template is rendered successfully",
			"warnings", len(theTemplate.Warnings()))

		for _, w := range theTemplate.Warnings() {
			Logger.Info("Warning: An issue occurred while rendering the Helm template",
				"issue", w)
		}
	}

	return theTemplate, nil
}
