package helm

import (
	autoscalingv2beta1 "k8s.io/api/autoscaling/v2beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (q *cachingQuery) HPAByName(name string) *autoscalingv2beta1.HorizontalPodAutoscaler {
	key := q.cacheKey(name, gvkHorizontalPodAutoscaler, nil)
	result := q.runQuery(key,
		func() interface{} {
			objects, err := q.template.GetObjects(
				NewHorizontalPodAutoscalerSelector(
					func(d *autoscalingv2beta1.HorizontalPodAutoscaler) bool {
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

	services := result.([]*autoscalingv2beta1.HorizontalPodAutoscaler)

	if len(services) == 0 {
		return nil
	}

	return services[0]
}

func (q *cachingQuery) HPAByLabels(labels map[string]string) []*autoscalingv2beta1.HorizontalPodAutoscaler {
	key := q.cacheKey(anything, gvkHorizontalPodAutoscaler, labels)
	result := q.runQuery(key,
		func() interface{} {
			objects, err := q.template.GetObjects(
				NewHorizontalPodAutoscalerSelector(
					func(d *autoscalingv2beta1.HorizontalPodAutoscaler) bool {
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

	return result.([]*autoscalingv2beta1.HorizontalPodAutoscaler)
}

func (q *cachingQuery) HPAByComponent(component string) *autoscalingv2beta1.HorizontalPodAutoscaler {
	hpas := q.HPAByLabels(map[string]string{
		appLabel: component,
	})
	if len(hpas) == 0 {
		return nil
	}

	return hpas[0]
}

func unsafeConvertHPAs(objects []runtime.Object) []*autoscalingv2beta1.HorizontalPodAutoscaler {
	hpas := make([]*autoscalingv2beta1.HorizontalPodAutoscaler, len(objects))
	for i, o := range objects {
		hpas[i] = o.(*autoscalingv2beta1.HorizontalPodAutoscaler)
	}

	return hpas
}
