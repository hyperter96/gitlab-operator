package helm

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (q *cachingQuery) ServiceByName(name string) *corev1.Service {
	key := q.cacheKey(name, gvkService, nil)
	result := q.runQuery(key,
		func() interface{} {
			objects, err := q.template.GetObjects(
				NewServiceSelector(
					func(d *corev1.Service) bool {
						return d.ObjectMeta.Name == name
					},
				),
			)

			if err != nil {
				return nil
			}

			return unsafeConvertServices(objects)
		},
	)

	services := result.([]*corev1.Service)

	if len(services) == 0 {
		return nil
	}

	return services[0]
}

func (q *cachingQuery) ServicesByLabels(labels map[string]string) []*corev1.Service {
	key := q.cacheKey(anything, gvkService, labels)
	result := q.runQuery(key,
		func() interface{} {
			objects, err := q.template.GetObjects(
				NewServiceSelector(
					func(d *corev1.Service) bool {
						return matchLabels(d.ObjectMeta.Labels, labels)
					},
				),
			)

			if err != nil {
				return nil
			}

			return unsafeConvertServices(objects)
		},
	)

	return result.([]*corev1.Service)
}

func (q *cachingQuery) ServiceByComponent(component string) *corev1.Service {
	services := q.ServicesByLabels(map[string]string{
		appLabel: component,
	})
	if len(services) == 0 {
		return nil
	}

	return services[0]
}

func unsafeConvertServices(objects []runtime.Object) []*corev1.Service {
	services := make([]*corev1.Service, len(objects))
	for i, o := range objects {
		services[i] = o.(*corev1.Service)
	}

	return services
}
