package helm

import (
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (q *cachingQuery) IngressesByLabels(labels map[string]string) []*networkingv1.Ingress {
	key := q.cacheKey(anything, gvkIngress, labels)
	result := q.runQuery(key,
		func() interface{} {
			objects, err := q.template.GetObjects(
				NewIngressSelector(
					func(d *networkingv1.Ingress) bool {
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

	return result.([]*networkingv1.Ingress)
}

func (q *cachingQuery) IngressByComponent(component string) *networkingv1.Ingress {
	ingresses := q.IngressesByLabels(map[string]string{
		appLabel: component,
	})
	if len(ingresses) == 0 {
		return nil
	}

	return ingresses[0]
}

func unsafeConvertIngresses(objects []runtime.Object) []*networkingv1.Ingress {
	ingresses := make([]*networkingv1.Ingress, len(objects))
	for i, o := range objects {
		ingresses[i] = o.(*networkingv1.Ingress)
	}

	return ingresses
}
