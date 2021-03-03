package helm

import (
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (q *cachingQuery) DeploymentByName(name string) *appsv1.Deployment {
	key := q.cacheKey(name, gvkDeployment, nil)
	result := q.runQuery(key,
		func() interface{} {
			objects, err := q.template.GetObjects(
				NewDeploymentSelector(
					func(d *appsv1.Deployment) bool {
						return d.ObjectMeta.Name == name
					},
				),
			)
			if err != nil {
				return nil
			}
			return unsafeConvertDeployments(objects)
		},
	)

	deployments := result.([]*appsv1.Deployment)

	if len(deployments) == 0 {
		return nil
	}
	return deployments[0]
}

func (q *cachingQuery) DeploymentsByLabels(labels map[string]string) []*appsv1.Deployment {
	key := q.cacheKey(anything, gvkDeployment, labels)
	result := q.runQuery(key,
		func() interface{} {
			objects, err := q.template.GetObjects(
				NewDeploymentSelector(
					func(d *appsv1.Deployment) bool {
						return matchLabels(d.ObjectMeta.Labels, labels)
					},
				),
			)
			if err != nil {
				return nil
			}
			return unsafeConvertDeployments(objects)
		},
	)
	return result.([]*appsv1.Deployment)
}

func (q *cachingQuery) DeploymentByComponent(component string) *appsv1.Deployment {
	deployments := q.DeploymentsByLabels(map[string]string{
		appLabel: component,
	})
	if len(deployments) == 0 {
		return nil
	}
	return deployments[0]
}

func unsafeConvertDeployments(objects []runtime.Object) []*appsv1.Deployment {
	deployments := make([]*appsv1.Deployment, len(objects))
	for i, o := range objects {
		deployments[i] = o.(*appsv1.Deployment)
	}
	return deployments
}
