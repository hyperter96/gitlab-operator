package gitlab

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
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
func RedisConfigMaps(adapter CustomResourceAdapter) []client.Object {
	template, err := GetTemplate(adapter)
	if err != nil {
		return []client.Object{} // WARNING: this should return an error instead.
	}

	result := template.Query().ObjectsByKindAndLabels(ConfigMapKind, map[string]string{
		"app": RedisComponentName,
	})

	for _, c := range result {
		updateCommonLabels(adapter.ReleaseName(), RedisComponentName, c.GetLabels())
	}

	return result
}

// RedisServices returns the Services of the Redis component.
func RedisServices(adapter CustomResourceAdapter) []client.Object {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: This should return an error instead.
	}

	results := template.Query().ObjectsByKindAndLabels(ServiceKind, map[string]string{
		"app": RedisComponentName,
	})

	for _, s := range results {
		updateCommonLabels(adapter.ReleaseName(), RedisComponentName, s.GetLabels())
	}

	return results
}

// RedisStatefulSet returns the Statefulset of the Redis component.
func RedisStatefulSet(adapter CustomResourceAdapter) client.Object {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: This should return an error instead.
	}

	result := template.Query().ObjectByKindAndComponent(StatefulSetKind, RedisComponentName)

	return result
}
