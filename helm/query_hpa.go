package helm

import (
	corev1 "k8s.io/api/core/v1"
  autoscalingv1 "k8s.io/api/autoscaling/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (q *cachingQuery) HPAByName(name string) *autoscalingv1.HorizontalPodAutoscaler {
	key := q.cacheKey(name, gvkHorizontalPodAutoscaler, nil)
	result := q.runQuery(key,
		func() interface{} {
			objects, err := q.template.GetObjects(
				NewHorizontalPodAutoscalerSelector(
					func(d *autoscalingv1.HorizontalPodAutoscaler) bool {
						return d.ObjectMeta.Name == name
					},
				),
			)
			if err != nil {
				return nil
			}
			return unsafeConvertHPAs(objects)
		},
	)

	services := result.([]*autoscalingv1.HorizontalPodAutoscaler)

	if len(services) == 0 {
		return nil
	}
	return services[0]
}

func (q *cachingQuery) HPAByLabels(labels map[string]string) []*autoscalingv1.HorizontalPodAutoscaler {
	key := q.cacheKey(anything, gvkHorizontalPodAutoscaler, labels)
	result := q.runQuery(key,
		func() interface{} {
			objects, err := q.template.GetObjects(
				NewHorizontalPodAutoscalerSelector(
					func(d *autoscalingv1.HorizontalPodAutoscaler) bool {
						return matchLabels(d.ObjectMeta.Labels, labels)
					},
				),
			)
			if err != nil {
				return nil
			}
			return unsafeConvertHPAs(objects)
		},
	)
	return result.([]*autoscalingv1.HorizontalPodAutoscaler)
}

func (q *cachingQuery) HPAByComponent(component string) *autoscalingv1.HorizontalPodAutoscaler {
	hpas := q.HpaByLabels(map[string]string{
		appLabel: component,
	})
	if len(hpas) == 0 {
		return nil
	}
	return hpas[0]
}

func unsafeConvertHPAs(objects []runtime.Object) []*autoscalingv1.HorizontalPodAutoscaler {
	hpas := make([]*autoscalingv1.HorizontalPodAutoscaler, len(objects))
	for i, o := range objects {
		hpas[i] = o.(*autoscalingv1.HorizontalPodAutoscaler)
	}
	return hpas
}
