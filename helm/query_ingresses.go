package helm

import (
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (q *cachingQuery) IngressesByLabels(labels map[string]string) []*extensionsv1beta1.Ingress {
	key := q.cacheKey(anything, gvkIngress, labels)
	result := q.runQuery(key,
		func() interface{} {
			objects, err := q.template.GetObjects(
				NewIngressSelector(
					func(d *extensionsv1beta1.Ingress) bool {
						return matchLabels(d.ObjectMeta.Labels, labels)
					},
				),
			)

			if err != nil {
				return nil
			}

			return unsafeConvertIngresses(objects)
		},
	)

	return result.([]*extensionsv1beta1.Ingress)
}

func (q *cachingQuery) IngressByComponent(component string) *extensionsv1beta1.Ingress {
	ingresses := q.IngressesByLabels(map[string]string{
		appLabel: component,
	})
	if len(ingresses) == 0 {
		return nil
	}

	return ingresses[0]
}

func unsafeConvertIngresses(objects []runtime.Object) []*extensionsv1beta1.Ingress {
	ingresses := make([]*extensionsv1beta1.Ingress, len(objects))
	for i, o := range objects {
		ingresses[i] = o.(*extensionsv1beta1.Ingress)
	}

	return ingresses
}
