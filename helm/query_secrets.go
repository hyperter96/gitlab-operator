package helm

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (q *cachingQuery) SecretByName(name string) *corev1.Secret {
	key := q.cacheKey(name, gvkSecret, nil)
	result := q.runQuery(key,
		func() interface{} {
			objects, err := q.template.GetObjects(
				NewSecretSelector(
					func(d *corev1.Secret) bool {
						return d.ObjectMeta.Name == name
					},
				),
			)

			if err != nil {
				return nil
			}

			return unsafeConvertSecrets(objects)
		},
	)

	secrets := result.([]*corev1.Secret)

	if len(secrets) == 0 {
		return nil
	}

	return secrets[0]
}

func (q *cachingQuery) SecretsByLabels(labels map[string]string) []*corev1.Secret {
	key := q.cacheKey(anything, gvkSecret, labels)
	result := q.runQuery(key,
		func() interface{} {
			objects, err := q.template.GetObjects(
				NewSecretSelector(
					func(d *corev1.Secret) bool {
						return matchLabels(d.ObjectMeta.Labels, labels)
					},
				),
			)

			if err != nil {
				return nil
			}

			return unsafeConvertSecrets(objects)
		},
	)

	return result.([]*corev1.Secret)
}

func unsafeConvertSecrets(objects []runtime.Object) []*corev1.Secret {
	secrets := make([]*corev1.Secret, len(objects))
	for i, o := range objects {
		secrets[i] = o.(*corev1.Secret)
	}

	return secrets
}
