package helm

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	accessor = meta.NewAccessor()
)

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
				return nil
			}

			return objects
		},
	)

	return result.([]runtime.Object)
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
