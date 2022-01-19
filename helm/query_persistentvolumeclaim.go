package helm

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (q *cachingQuery) PersistentVolumeClaimByName(name string) *corev1.PersistentVolumeClaim {
	key := q.cacheKey(name, gvkPersistentVolumeClaim, nil)
	result := q.runQuery(key,
		func() interface{} {
			objects, err := q.template.GetObjects(
				NewPersistentVolumeClaimSelector(
					func(d *corev1.PersistentVolumeClaim) bool {
						return d.ObjectMeta.Name == name
					},
				),
			)

			if err != nil {
				return nil
			}

			return unsafeConvertPersistentVolumeClaims(objects)
		},
	)

	persistentVolumeClaims := result.([]*corev1.PersistentVolumeClaim)

	if len(persistentVolumeClaims) == 0 {
		return nil
	}

	return persistentVolumeClaims[0]
}

func unsafeConvertPersistentVolumeClaims(objects []runtime.Object) []*corev1.PersistentVolumeClaim {
	persistentVolumeClaims := make([]*corev1.PersistentVolumeClaim, len(objects))
	for i, o := range objects {
		persistentVolumeClaims[i] = o.(*corev1.PersistentVolumeClaim)
	}

	return persistentVolumeClaims
}
