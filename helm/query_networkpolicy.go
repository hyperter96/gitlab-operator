package helm

import (
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (q *cachingQuery) NetworkPolicyByName(name string) *networkingv1.NetworkPolicy {
	key := q.cacheKey(name, gvkNetworkPolicy, nil)
	result := q.runQuery(key,
		func() interface{} {
			objects, err := q.template.GetObjects(
				NewNetworkPolicySelector(
					func(d *networkingv1.NetworkPolicy) bool {
						return d.ObjectMeta.Name == name
					},
				),
			)
			if err != nil {
				return nil
			}
			return unsafeConvertNetworkPolicies(objects)
		},
	)

	policies := result.([]*networkingv1.NetworkPolicy)

	if len(policies) == 0 {
		return nil
	}

	return policies[0]
}

func (q *cachingQuery) NetworkPolicyByLabels(labels map[string]string) []*networkingv1.NetworkPolicy {
	key := q.cacheKey(anything, gvkNetworkPolicy, labels)
	result := q.runQuery(key,
		func() interface{} {
			objects, err := q.template.GetObjects(
				NewNetworkPolicySelector(
					func(d *networkingv1.NetworkPolicy) bool {
						return matchLabels(d.ObjectMeta.Labels, labels)
					},
				),
			)
			if err != nil {
				return nil
			}
			return unsafeConvertNetworkPolicies(objects)
		},
	)

	return result.([]*networkingv1.NetworkPolicy)
}

func (q *cachingQuery) NetworkPolicyByComponent(component string) *networkingv1.NetworkPolicy {
	policies := q.NetworkPolicyByLabels(map[string]string{
		appLabel: component,
	})
	if len(policies) == 0 {
		return nil
	}

	return policies[0]
}

func unsafeConvertNetworkPolicies(objects []runtime.Object) []*networkingv1.NetworkPolicy {
	policies := make([]*networkingv1.NetworkPolicy, len(objects))
	for i, o := range objects {
		policies[i] = o.(*networkingv1.NetworkPolicy)
	}

	return policies
}
