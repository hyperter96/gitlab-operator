package gitlab

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	GlobalGitalyEnabled  = "global.gitaly.enabled"
	gitalyEnabledDefault = true
)

// GitalyEnabled returns `true` if enabled and `false` if not.
func GitalyEnabled(adapter CustomResourceAdapter) bool {
	return adapter.Values().GetBool(GlobalGitalyEnabled, gitalyEnabledDefault)
}

// GitalyStatefulSet returns the StatefulSet of Gitaly component.
func GitalyStatefulSet(adapter CustomResourceAdapter) client.Object {
	template, err := GetTemplate(adapter)

	if err != nil {
		return nil // WARNING: This should return an error instead.
	}

	result := template.Query().ObjectByKindAndComponent(StatefulSetKind, GitalyComponentName)

	return result
}

// GitalyConfigMap returns the ConfigMap of Gitaly component.
func GitalyConfigMap(adapter CustomResourceAdapter) client.Object {
	template, err := GetTemplate(adapter)

	if err != nil {
		return nil // WARNING: This should return an error instead.
	}

	result := template.Query().ObjectByKindAndComponent(ConfigMapKind, GitalyComponentName)

	return result
}

// GitalyService returns the Service of GitLab Shell component.
func GitalyService(adapter CustomResourceAdapter) client.Object {
	template, err := GetTemplate(adapter)

	if err != nil {
		return nil // WARNING: This should return an error instead.
	}

	result := template.Query().ObjectByKindAndComponent(ServiceKind, GitalyComponentName)

	return result
}
