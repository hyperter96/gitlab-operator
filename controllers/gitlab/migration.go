package gitlab

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
)

const (
	gitlabMigrationsEnabled = "gitlab.migrations.enabled"
)

// MigrationsEnabled returns `true` if enabled and `false` if not.
func MigrationsEnabled(adapter CustomResourceAdapter) bool {
	return adapter.Values().GetBool(gitlabMigrationsEnabled)
}

// MigrationsConfigMap returns the ConfigMaps of Migrations component.
func MigrationsConfigMap(adapter CustomResourceAdapter, template helm.Template) client.Object {
	return template.Query().ObjectByKindAndName(ConfigMapKind,
		fmt.Sprintf("%s-%s", adapter.ReleaseName(), MigrationsComponentName))
}

// MigrationsJob returns the Job for Migrations component.
func MigrationsJob(adapter CustomResourceAdapter, template helm.Template) (client.Object, error) {
	result := template.Query().ObjectByKindAndComponent(JobKind, MigrationsComponentName)
	result.SetName(nameWithHashSuffix(result.GetName(), adapter, 3))

	return result, nil
}
