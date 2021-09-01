package helm

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (q *cachingQuery) ServiceAccountByName(name string) *corev1.ServiceAccount {
	key := q.cacheKey(name, gvkServiceAccount, nil)
	result := q.runQuery(key,
		func() interface{} {
			objects, err := q.template.GetObjects(
				NewServiceAccountSelector(
					func(d *corev1.ServiceAccount) bool {
						return d.ObjectMeta.Name == name
					},
				),
			)
			if err != nil {
				return nil
			}
			return unsafeConvertServiceAccounts(objects)
		},
	)

	accounts := result.([]*corev1.ServiceAccount)

	if len(accounts) == 0 {
		return nil
	}
	return accounts[0]
}

func (q *cachingQuery) ServiceAccountByLabels(labels map[string]string) []*corev1.ServiceAccount {
	key := q.cacheKey(anything, gvkServiceAccount, labels)
	result := q.runQuery(key,
		func() interface{} {
			objects, err := q.template.GetObjects(
				NewServiceAccountSelector(
					func(d *corev1.ServiceAccount) bool {
						return matchLabels(d.ObjectMeta.Labels, labels)
					},
				),
			)
			if err != nil {
				return nil
			}
			return unsafeConvertServiceAccounts(objects)
		},
	)
	return result.([]*corev1.ServiceAccount)
}

func (q *cachingQuery) ServiceAccountByComponent(component string) *corev1.ServiceAccount {
	accounts := q.ServiceAccountByLabels(map[string]string{
		appLabel: component,
	})
	if len(accounts) == 0 {
		return nil
	}
	return accounts[0]
}

func unsafeConvertServiceAccounts(objects []runtime.Object) []*corev1.ServiceAccount {
	accounts := make([]*corev1.ServiceAccount, len(objects))
	for i, o := range objects {
		accounts[i] = o.(*corev1.ServiceAccount)
	}
	return accounts
}
