package helm

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (q *cachingQuery) ConfigMapByName(name string) *corev1.ConfigMap {
	key := q.cacheKey(name, gvkConfigMap, nil)
	result := q.runQuery(key,
		func() interface{} {
			objects, err := q.template.GetObjects(
				NewConfigMapSelector(
					func(d *corev1.ConfigMap) bool {
						return d.ObjectMeta.Name == name
					},
				),
			)

			if err != nil {
				return nil
			}

			return unsafeConvertConfigMaps(objects)
		},
	)

	configMaps := result.([]*corev1.ConfigMap)

	if len(configMaps) == 0 {
		return nil
	}

	return configMaps[0]
}

func (q *cachingQuery) ConfigMapsByLabels(labels map[string]string) []*corev1.ConfigMap {
	key := q.cacheKey(anything, gvkConfigMap, labels)
	result := q.runQuery(key,
		func() interface{} {
			objects, err := q.template.GetObjects(
				NewConfigMapSelector(
					func(d *corev1.ConfigMap) bool {
						return matchLabels(d.ObjectMeta.Labels, labels)
					},
				),
			)

			if err != nil {
				return nil
			}

			return unsafeConvertConfigMaps(objects)
		},
	)

	return result.([]*corev1.ConfigMap)
}

func (q *cachingQuery) ConfigMapByComponent(component string) *corev1.ConfigMap {
	configMaps := q.ConfigMapsByLabels(map[string]string{
		appLabel: component,
	})

	if len(configMaps) == 0 {
		return nil
	}

	return configMaps[0]
}

func unsafeConvertConfigMaps(objects []runtime.Object) []*corev1.ConfigMap {
	configMaps := make([]*corev1.ConfigMap, len(objects))
	for i, o := range objects {
		configMaps[i] = o.(*corev1.ConfigMap)
	}

	return configMaps
}
