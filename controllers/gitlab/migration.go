package gitlab

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/gitlab"
)

// MigrationsConfigMap returns the ConfigMaps of Migrations component.
func MigrationsConfigMap(adapter gitlab.Adapter, template helm.Template) client.Object {
	return template.Query().ObjectByKindAndName(ConfigMapKind,
		fmt.Sprintf("%s-%s", adapter.ReleaseName(), MigrationsComponentName))
}

// MigrationsJob returns the Job for Migrations component.
func MigrationsJob(adapter gitlab.Adapter, template helm.Template) (client.Object, error) {
	result := template.Query().ObjectByKindAndComponent(JobKind, MigrationsComponentName)
	result.SetName(nameWithHashSuffix(result.GetName(), adapter, 3))

	return result, nil
}
