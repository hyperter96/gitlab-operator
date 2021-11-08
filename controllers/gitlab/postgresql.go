package gitlab

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	postgresInstall        = "postgresql.install"
	postgresEnabledDefault = true
)

// PostgresEnabled returns `true` if `PostgreSQL` is enabled, and `false` if not.
func PostgresEnabled(adapter CustomResourceAdapter) bool {
	enabled, _ := GetBoolValue(adapter.Values(), postgresInstall, postgresEnabledDefault)

	return enabled
}

// PostgresServices returns the Services of the Postgres component.
func PostgresServices(adapter CustomResourceAdapter) []*corev1.Service {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: This should return an error instead.
	}

	results := template.Query().ServicesByLabels(map[string]string{
		"app": PostgresComponentName,
	})

	for _, s := range results {
		updateCommonLabels(adapter.ReleaseName(), PostgresComponentName, s.ObjectMeta.Labels)

		// Temporary fix: patch in the namespace because the version of the PostgreSQL chart
		// we use does not specify `namespace` in the template.
		s.ObjectMeta.Namespace = adapter.Namespace()
	}

	return results
}

// PostgresStatefulSet returns the StatefulSet of the PostgreSQL component.
func PostgresStatefulSet(adapter CustomResourceAdapter) *appsv1.StatefulSet {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: This should return an error instead.
	}

	result := template.Query().StatefulSetByComponent(PostgresComponentName)

	// Temporary fix: patch in the namespace because the version of the PostgreSQL chart
	// we use does not specify `namespace` in the template.
	result.ObjectMeta.Namespace = adapter.Namespace()

	updateCommonLabels(adapter.ReleaseName(), PostgresComponentName, result.ObjectMeta.Labels)
	updateCommonLabels(adapter.ReleaseName(), PostgresComponentName, result.Spec.Template.ObjectMeta.Labels)

	return result
}

// PostgresConfigMap returns the ConfigMap of the PostgreSQL component.
func PostgresConfigMap(adapter CustomResourceAdapter) *corev1.ConfigMap {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	initDBConfigMap := template.Query().ConfigMapByName(
		fmt.Sprintf("%s-postgresql-init-db", adapter.ReleaseName()))

	updateCommonLabels(adapter.ReleaseName(), PostgresComponentName, initDBConfigMap.ObjectMeta.Labels)

	return initDBConfigMap
}
