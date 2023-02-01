package gitlab

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/gitlab"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/support"
)

// MigrationsConfigMap returns the ConfigMaps of Migrations component.
func MigrationsConfigMap(adapter gitlab.Adapter, template helm.Template) client.Object {
	return template.Query().ObjectByKindAndName(ConfigMapKind,
		fmt.Sprintf("%s-%s", adapter.ReleaseName(), MigrationsComponentName))
}

// MigrationsJob returns the Job for Migrations component.
func MigrationsJob(adapter gitlab.Adapter, template helm.Template) (client.Object, error) {
	result := template.Query().ObjectByKindAndComponent(JobKind, MigrationsComponentName)

	nameWithSuffix, err := support.NameWithHashSuffix(result.GetName(), adapter.Hash(), 5)
	if err != nil {
		return result, err
	}

	result.SetName(nameWithSuffix)

	return result, nil
}
