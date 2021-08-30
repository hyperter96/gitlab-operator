package helm

import (
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (q *cachingQuery) StatefulSetByName(name string) *appsv1.StatefulSet {
	key := q.cacheKey(name, gvkStatefulSet, nil)
	result := q.runQuery(key,
		func() interface{} {
			objects, err := q.template.GetObjects(
				NewStatefulSetSelector(
					func(d *appsv1.StatefulSet) bool {
						return d.ObjectMeta.Name == name
					},
				),
			)

			if err != nil {
				return nil
			}

			return unsafeConvertStatefulSets(objects)
		},
	)

	statefulSets := result.([]*appsv1.StatefulSet)

	if len(statefulSets) == 0 {
		return nil
	}

	return statefulSets[0]
}

func (q *cachingQuery) StatefulSetsByLabels(labels map[string]string) []*appsv1.StatefulSet {
	key := q.cacheKey(anything, gvkStatefulSet, labels)
	result := q.runQuery(key,
		func() interface{} {
			objects, err := q.template.GetObjects(
				NewStatefulSetSelector(
					func(d *appsv1.StatefulSet) bool {
						return matchLabels(d.ObjectMeta.Labels, labels)
					},
				),
			)

			if err != nil {
				return nil
			}

			return unsafeConvertStatefulSets(objects)
		},
	)

	return result.([]*appsv1.StatefulSet)
}

func (q *cachingQuery) StatefulSetByComponent(component string) *appsv1.StatefulSet {
	statefulSets := q.StatefulSetsByLabels(map[string]string{
		appLabel: component,
	})
	if len(statefulSets) == 0 {
		return nil
	}

	return statefulSets[0]
}

func unsafeConvertStatefulSets(objects []runtime.Object) []*appsv1.StatefulSet {
	statefulSets := make([]*appsv1.StatefulSet, len(objects))
	for i, o := range objects {
		statefulSets[i] = o.(*appsv1.StatefulSet)
	}

	return statefulSets
}
