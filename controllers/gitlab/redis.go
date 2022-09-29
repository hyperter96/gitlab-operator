package gitlab

import (
	"sigs.k8s.io/controller-runtime/pkg/client"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/gitlab"
)

// RedisConfigMaps returns the ConfigMaps of the Redis component.
func RedisConfigMaps(adapter gitlab.Adapter, template helm.Template) []client.Object {
	result := template.Query().ObjectsByKindAndLabels(ConfigMapKind, map[string]string{
		"app": RedisComponentName,
	})

	for _, c := range result {
		updateCommonLabels(adapter.ReleaseName(), RedisComponentName, c.GetLabels())
	}

	return result
}

// RedisServices returns the Services of the Redis component.
func RedisServices(adapter gitlab.Adapter, template helm.Template) []client.Object {
	results := template.Query().ObjectsByKindAndLabels(ServiceKind, map[string]string{
		"app": RedisComponentName,
	})

	for _, s := range results {
		updateCommonLabels(adapter.ReleaseName(), RedisComponentName, s.GetLabels())
	}

	return results
}

// RedisStatefulSet returns the Statefulset of the Redis component.
func RedisStatefulSet(template helm.Template) client.Object {
	return template.Query().ObjectByKindAndComponent(StatefulSetKind, RedisComponentName)
}
