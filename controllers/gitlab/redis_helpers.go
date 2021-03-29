package gitlab

import (
	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/helpers"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// RedisConfigMaps returns the ConfigMaps of the Redis component.
func RedisConfigMaps(adapter helpers.CustomResourceAdapter) []*corev1.ConfigMap {
	template, err := helpers.GetTemplate(adapter)
	if err != nil {
		return []*corev1.ConfigMap{} // WARNING: this should return an error instead.
	}

	result := template.Query().ConfigMapsByLabels(map[string]string{
		"app": RedisComponentName,
	})

	for _, c := range result {
		updateCommonLabels(adapter.ReleaseName(), RedisComponentName, &c.ObjectMeta.Labels)
	}

	return result
}

// RedisServices returns the Services of the Redis component.
func RedisServices(adapter helpers.CustomResourceAdapter) []*corev1.Service {
	template, err := helpers.GetTemplate(adapter)
	if err != nil {
		return nil
		/* WARNING: This should return an error instead. */
	}

	results := template.Query().ServicesByLabels(map[string]string{
		"app": RedisComponentName,
	})

	for _, s := range results {
		updateCommonLabels(adapter.ReleaseName(), RedisComponentName, &s.ObjectMeta.Labels)
	}

	return results
}

// RedisStatefulSet returns the Statefulset of the Redis component.
func RedisStatefulSet(adapter helpers.CustomResourceAdapter) *appsv1.StatefulSet {
	template, err := helpers.GetTemplate(adapter)
	if err != nil {
		return nil
		/* WARNING: This should return an error instead. */
	}

	result := template.Query().StatefulSetByComponent(RedisComponentName)

	return result
}
