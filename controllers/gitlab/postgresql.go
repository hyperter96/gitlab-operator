package gitlab

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
)

const (
	postgresInstall = "postgresql.install"
)

// PostgresEnabled returns `true` if `PostgreSQL` is enabled, and `false` if not.
func PostgresEnabled(adapter CustomResourceAdapter) bool {
	return adapter.Values().GetBool(postgresInstall)
}

// PostgresServices returns the Services of the Postgres component.
func PostgresServices(adapter CustomResourceAdapter, template helm.Template) []client.Object {
	results := template.Query().ObjectsByKindAndLabels(ServiceKind, map[string]string{
		"app": PostgresComponentName,
	})

	for _, s := range results {
		updateCommonLabels(adapter.ReleaseName(), PostgresComponentName, s.GetLabels())

		// Temporary fix: patch in the namespace because the version of the PostgreSQL chart
		// we use does not specify `namespace` in the template.
		s.SetNamespace(adapter.Namespace())
	}

	return results
}

// PostgresStatefulSet returns the StatefulSet of the PostgreSQL component.
func PostgresStatefulSet(adapter CustomResourceAdapter, template helm.Template) client.Object {
	result := template.Query().ObjectByKindAndComponent(StatefulSetKind, PostgresComponentName)

	// Temporary fix: patch in the namespace because the version of the PostgreSQL chart
	// we use does not specify `namespace` in the template.
	result.SetNamespace(adapter.Namespace())

	updateCommonLabels(adapter.ReleaseName(), PostgresComponentName, result.GetLabels())

	// Attention: Type Assertion: StatefulSet.Spec is needed
	updateCommonLabels(adapter.ReleaseName(), PostgresComponentName, result.(*appsv1.StatefulSet).Spec.Template.ObjectMeta.Labels)

	return result
}

// PostgresConfigMap returns the ConfigMap of the PostgreSQL component.
func PostgresConfigMap(adapter CustomResourceAdapter, template helm.Template) client.Object {
	initDBConfigMap := template.Query().ObjectByKindAndName(ConfigMapKind,
		fmt.Sprintf("%s-postgresql-init-db", adapter.ReleaseName()))

	updateCommonLabels(adapter.ReleaseName(), PostgresComponentName, initDBConfigMap.GetLabels())

	return initDBConfigMap
}
