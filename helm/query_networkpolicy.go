package helm

import (
	corev1 "k8s.io/api/core/v1"
  networkpolicyv1 "k8s.io/api/networking.k8s.io/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (q *cachingQuery) NetworkPolicyByName(name string) *networkpolicyv1.NetworkPolicy {
	key := q.cacheKey(name, gvkNetworkPolicy, nil)
	result := q.runQuery(key,
		func() interface{} {
			objects, err := q.template.GetObjects(
				NewNetworkPolicySelector(
					func(d *networkpolicyv1.NetworkPolicy) bool {
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

	services := result.([]*networkpolicyv1.NetworkPolicy)

	if len(services) == 0 {
		return nil
	}
	return services[0]
}

func (q *cachingQuery) NetworkPolicyByLabels(labels map[string]string) []*networkpolicyv1.NetworkPolicy {
	key := q.cacheKey(anything, gvkNetworkPolicy, labels)
	result := q.runQuery(key,
		func() interface{} {
			objects, err := q.template.GetObjects(
				NewNetworkPolicySelector(
					func(d *networkpolicyv1.NetworkPolicy) bool {
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
	return result.([]*networkpolicyv1.NetworkPolicy)
}

func (q *cachingQuery) NetworkPolicyByComponent(component string) *networkpolicyv1.NetworkPolicy {
	networkPolicies := q.NetworkPolicyByLabels(map[string]string{
		appLabel: component,
	})
	if len(networkPolicies) == 0 {
		return nil
	}
	return networkPolicies[0]
}

func unsafeConvertNetworkPolicies(objects []runtime.Object) []*networkpolicyv1.NetworkPolicy {
	networkPolicies := make([]*networkpolicyv1.NetworkPolicy, len(objects))
	for i, o := range objects {
		networkPolicies[i] = o.(*networkpolicyv1.NetworkPolicy)
	}
	return networkPolicies
}
