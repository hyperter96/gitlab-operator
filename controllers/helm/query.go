package helm

import (
	"fmt"
	"sync"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Query provides access methods to query Helm templates.
type Query interface {
	// Template returns the attached template that this interface queries.
	Template() Template

	// ObjectsByKind returns all objects that match the kind specifier. Type specifier can be in the form of
	// Kind, Kind.group, Kind.version.group.
	ObjectsByKind(kindArg string) []runtime.Object

	// ObjectByKindAndName returns the object that match the kind specifier and has the provided name.
	ObjectByKindAndName(kindArg, name string) runtime.Object

	// ObjectByKindAndLabels returns the all objects that match the kind specifier and have the labels.
	ObjectsByKindAndLabels(kindArg string, labels map[string]string) []runtime.Object

	// DeploymentsByLabels lists all Deployments that match the labels.
	DeploymentsByLabels(labels map[string]string) []*appsv1.Deployment

	// DeploymentByComponent returns the Deployment for a specific component.
	DeploymentByComponent(component string) *appsv1.Deployment

	// Reset clears the query cache when applicable.
	Reset()
}

type cachingQuery struct {
	template Template
	cache    map[string]interface{}
	lock     *sync.Mutex
}

func newQuery(t Template) Query {
	return &cachingQuery{
		template: t,
		cache:    make(map[string]interface{}),
		lock:     &sync.Mutex{},
	}
}

const (
	anything      = "*"
	gvkDeployment = "Deployment.v1.apps"
)

var (
	accessor = meta.NewAccessor()
)

func (q *cachingQuery) cacheKey(nameOrComponent, gvk string, labels map[string]string) string {
	return fmt.Sprintf("%s.%s[%s]", nameOrComponent, gvk, labels)
}

func (q *cachingQuery) readCache(key string) interface{} {
	q.lock.Lock()
	defer q.lock.Unlock()

	if result, ok := q.cache[key]; ok {
		return result
	}
	return nil
}

func (q *cachingQuery) updateCache(key string, objects interface{}) {
	if objects == nil {
		return
	}

	q.lock.Lock()
	defer q.lock.Unlock()

	q.cache[key] = objects
}

func (q *cachingQuery) clearCache() {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.cache = make(map[string]interface{})
}

func (q *cachingQuery) runQuery(key string, query func() interface{}) interface{} {
	result := q.readCache(key)

	if result == nil {
		result = query()
		q.updateCache(key, result)
	}

	return result
}

func (q *cachingQuery) queryObjectsWithKindArg(key, kindArg string, selector ObjectSelector) []runtime.Object {
	gvk, gk := schema.ParseKindArg(kindArg)
	result := q.runQuery(key,
		func() interface{} {
			objects, err := q.template.GetObjects(
				func(obj runtime.Object) bool {
					return matchParsedKindArg(obj, gvk, &gk) && selector(obj)
				},
			)
			if err != nil {
				Logger.Error(err, "Unexpected error while querying kind", "kindArg", kindArg)
				return nil
			}
			return objects
		},
	)
	return result.([]runtime.Object)

}

func (q *cachingQuery) Template() Template {
	return q.template
}

func (q *cachingQuery) ObjectsByKind(kindArg string) []runtime.Object {
	key := q.cacheKey(anything, fmt.Sprintf("%s?", kindArg), nil)
	return q.queryObjectsWithKindArg(key, kindArg, TrueSelector)
}

func (q *cachingQuery) ObjectByKindAndName(kindArg, name string) runtime.Object {
	key := q.cacheKey(name, fmt.Sprintf("%s?", kindArg), nil)
	objects := q.queryObjectsWithKindArg(key, kindArg, func(obj runtime.Object) bool {
		objName, err := accessor.Name(obj)
		return err == nil && objName == name
	})
	if len(objects) == 0 {
		return nil
	}
	return objects[0]
}

func (q *cachingQuery) ObjectsByKindAndLabels(kindArg string, labels map[string]string) []runtime.Object {
	key := q.cacheKey(anything, fmt.Sprintf("%s?", kindArg), labels)
	return q.queryObjectsWithKindArg(key, kindArg, func(obj runtime.Object) bool {
		objLabels, err := accessor.Labels(obj)
		return err == nil && matchLabels(objLabels, labels)
	})
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
				Logger.Error(err, "Unexpected error while querying Deployments", "labels", labels)
				return nil
			}
			return unsafeConvertDeployments(objects)
		},
	)
	return (result).([]*appsv1.Deployment)
}

func (q *cachingQuery) DeploymentByComponent(component string) *appsv1.Deployment {
	deployments := q.DeploymentsByLabels(map[string]string{
		"app": component,
	})
	if len(deployments) > 0 {
		return deployments[0]
	}
	return nil
}

func (q *cachingQuery) Reset() {
	q.clearCache()
}

func unsafeConvertDeployments(objects []runtime.Object) []*appsv1.Deployment {
	deployments := make([]*appsv1.Deployment, len(objects))
	for i, o := range objects {
		deployments[i] = o.(*appsv1.Deployment)
	}
	return deployments
}

func matchLabels(oLabels, qLabels map[string]string) bool {
	for k, v := range qLabels {
		if w, ok := oLabels[k]; ok && v != w {
			return false
		}
	}
	return true
}

func matchParsedKindArg(object runtime.Object, qGVK *schema.GroupVersionKind, qGK *schema.GroupKind) bool {
	oKind, err := accessor.Kind(object)
	if err != nil {
		return false
	}

	oAPIVersion, err := accessor.APIVersion(object)
	if err != nil {
		return false
	}

	oGV, err := schema.ParseGroupVersion(oAPIVersion)
	if err != nil {
		return false
	}

	oGVK := oGV.WithKind(oKind)
	result := false

	if qGVK != nil {
		result = qGVK.Kind == oGVK.Kind &&
			(qGVK.Group == "" || qGVK.Group == oGVK.Group) &&
			(qGVK.Version == "" || qGVK.Version == oGVK.Version)
	}

	if result {
		return true
	}

	return qGK.Kind == oGVK.Kind &&
		(qGK.Group == "" || qGK.Group == oGVK.Group || qGK.Group == oGVK.Version)
}

// CacheBackdoor is used by test cases.
func CacheBackdoor(q Query) *map[string]interface{} {
	if cq, ok := (q).(*cachingQuery); ok {
		return &cq.cache
	}
	return nil
}
