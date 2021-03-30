package gitlab

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// GitalyStatefulSet returns the StatefulSet of Gitaly component.
func GitalyStatefulSet(adapter CustomResourceAdapter) *appsv1.StatefulSet {

	template, err := GetTemplate(adapter)

	if err != nil {
		return nil
		/* WARNING: This should return an error instead. */
	}

	result := template.Query().StatefulSetByComponent(GitalyComponentName)

	return result
}

// GitalyConfigMap returns the ConfigMap of Gitaly component.
func GitalyConfigMap(adapter CustomResourceAdapter) *corev1.ConfigMap {
	template, err := GetTemplate(adapter)

	if err != nil {
		return nil
		/* WARNING: This should return an error instead. */
	}

	result := template.Query().ConfigMapByComponent(GitalyComponentName)

	return result
}

// GitalyService returns the Service of GitLab Shell component.
func GitalyService(adapter CustomResourceAdapter) *corev1.Service {
	template, err := GetTemplate(adapter)

	if err != nil {
		return nil
		/* WARNING: This should return an error instead. */
	}

	result := template.Query().ServiceByComponent(GitalyComponentName)

	return result
}
