package gitlab

import (
	"sigs.k8s.io/controller-runtime/pkg/client"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
)

const (
	redisInstall        = "redis.install"
	redisEnabledDefault = true
)

// RedisEnabled returns `true` if Redis is enabled, and `false` if not.
func RedisEnabled(adapter CustomResourceAdapter) bool {
	return adapter.Values().GetBool(redisInstall, redisEnabledDefault)
}

// RedisConfigMaps returns the ConfigMaps of the Redis component.
func RedisConfigMaps(adapter CustomResourceAdapter, template helm.Template) []client.Object {
	result := template.Query().ObjectsByKindAndLabels(ConfigMapKind, map[string]string{
		"app": RedisComponentName,
	})

	for _, c := range result {
		updateCommonLabels(adapter.ReleaseName(), RedisComponentName, c.GetLabels())
	}

	return result
}

// RedisServices returns the Services of the Redis component.
func RedisServices(adapter CustomResourceAdapter, template helm.Template) []client.Object {
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
