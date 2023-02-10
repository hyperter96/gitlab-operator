package gitlab

import (
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/gitlab"
)

// PostgresServices returns the Services of the Postgres component.
func PostgresServices(adapter gitlab.Adapter, template helm.Template) []client.Object {
	componentName := PostgresComponentName(adapter)
	results := template.Query().ObjectsByKindAndLabels(ServiceKind, map[string]string{
		"app": componentName,
	})

	for _, s := range results {
		updateCommonLabels(adapter.ReleaseName(), componentName, s.GetLabels())

		// Temporary fix: patch in the namespace because the version of the PostgreSQL chart
		// we use does not specify `namespace` in the template.
		s.SetNamespace(adapter.Name().Namespace)
	}

	return results
}

// PostgresStatefulSet returns the StatefulSet of the PostgreSQL component.
func PostgresStatefulSet(adapter gitlab.Adapter, template helm.Template) client.Object {
	componentName := PostgresComponentName(adapter)
	result := template.Query().ObjectByKindAndComponent(StatefulSetKind, componentName)

	// Temporary fix: patch in the namespace because the version of the PostgreSQL chart
	// we use does not specify `namespace` in the template.
	result.SetNamespace(adapter.Name().Namespace)

	updateCommonLabels(adapter.ReleaseName(), componentName, result.GetLabels())

	// Attention: Type Assertion: StatefulSet.Spec is needed
	updateCommonLabels(adapter.ReleaseName(), componentName, result.(*appsv1.StatefulSet).Spec.Template.ObjectMeta.Labels)

	return result
}

// PostgresConfigMap returns the ConfigMap of the PostgreSQL component.
func PostgresConfigMap(adapter gitlab.Adapter, template helm.Template) client.Object {
	// InitDB Configmap's name currently does not respect name overrides
	initDBConfigMap := template.Query().ObjectByKindAndName(ConfigMapKind,
		fmt.Sprintf("%s-%s-init-db", adapter.ReleaseName(), DefaultPostgresComponentName))

	updateCommonLabels(adapter.ReleaseName(), PostgresComponentName(adapter), initDBConfigMap.GetLabels())

	return initDBConfigMap
}

// PostgresService returns the common Service of the PostgreSQL component.
func PostgresService(adapter gitlab.Adapter, template helm.Template) client.Object {
	pgServices := PostgresServices(adapter, template)

	for _, s := range pgServices {
		if !strings.HasSuffix(s.GetName(), "-headless") &&
			!strings.HasSuffix(s.GetName(), "-metrics") &&
			!strings.HasSuffix(s.GetName(), "-read") {
			return s
		}
	}

	return nil
}
