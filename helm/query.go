package helm

import (
	"fmt"
	"sync"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
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

	// ConfigMapByName returns the ConfigMap with the specified name.
	ConfigMapByName(name string) *corev1.ConfigMap

	// ConfigMapsByLabels lists all ConfigMaps that match the labels.
	ConfigMapsByLabels(labels map[string]string) []*corev1.ConfigMap

	// ConfigMapByComponent lists all ConfigMaps for a specific component.
	ConfigMapByComponent(component string) *corev1.ConfigMap

	// JobByName returns the Job with the specified name.
	JobByName(name string) *batchv1.Job

	// JobsByLabels lists all Jobs that match the labels.
	JobsByLabels(labels map[string]string) []*batchv1.Job

	// JobByComponent lists all Jobs for a specific component.
	JobByComponent(component string) *batchv1.Job

	// SecretByName returns the Secret with the specified name.
	SecretByName(name string) *corev1.Secret

	// SecretByLabels lists all Secrets that match the labels.
	SecretsByLabels(labels map[string]string) []*corev1.Secret

	// DeploymentByName returns the Deployment with the specified name.
	DeploymentByName(name string) *appsv1.Deployment

	// DeploymentsByLabels lists all Deployments that match the labels.
	DeploymentsByLabels(labels map[string]string) []*appsv1.Deployment

	// DeploymentByComponent returns the Deployment for a specific component.
	DeploymentByComponent(component string) *appsv1.Deployment

	// StatefulSetByName returns the StatefulSet with the specified name.
	StatefulSetByName(name string) *appsv1.StatefulSet

	// StatefulSetsByLabels lists all StatefulSets that match the labels.
	StatefulSetsByLabels(labels map[string]string) []*appsv1.StatefulSet

	// StatefulSetByComponent returns the StatefulSet for a specific component.
	StatefulSetByComponent(component string) *appsv1.StatefulSet

	// ServiceByName returns the Service with the specified name.
	ServiceByName(name string) *corev1.Service

	// ServicesByLabels lists all Services that match the labels.
	ServicesByLabels(labels map[string]string) []*corev1.Service

	// ServiceByComponent returns the Service for a specific component.
	ServiceByComponent(component string) *corev1.Service

	// Reset clears the query cache when applicable.
	Reset()
}

type cachingQuery struct {
	template Template
	cache    map[string]interface{}
	locker   sync.Locker
}

func newQuery(t Template) Query {
	return &cachingQuery{
		template: t,
		cache:    make(map[string]interface{}),
		locker:   &sync.Mutex{},
	}
}

const (
	anything       = "*"
	gvkDeployment  = "Deployment.v1.apps"
	gvkStatefulSet = "StatefulSet.v1.apps"
	gvkJob         = "Job.v1.batch"
	gvkConfigMap   = "ConfigMap.v1.core"
	gvkSecret      = "Secret.v1.core"
	gvkService     = "Service.v1.core"
)

var (
	accessor = meta.NewAccessor()
)

func (q *cachingQuery) cacheKey(nameOrComponent, gvk string, labels map[string]string) string {
	return fmt.Sprintf("%s.%s[%s]", nameOrComponent, gvk, labels)
}

func (q *cachingQuery) readCache(key string) interface{} {
	q.locker.Lock()
	defer q.locker.Unlock()

	if result, ok := q.cache[key]; ok {
		return result
	}
	return nil
}

func (q *cachingQuery) updateCache(key string, objects interface{}) {
	if objects == nil {
		return
	}

	q.locker.Lock()
	defer q.locker.Unlock()

	q.cache[key] = objects
}

func (q *cachingQuery) clearCache() {
	q.locker.Lock()
	defer q.locker.Unlock()

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

func (q *cachingQuery) ConfigMapByName(name string) *corev1.ConfigMap {
	key := q.cacheKey(name, gvkConfigMap, nil)
	result := q.runQuery(key,
		func() interface{} {
			objects, err := q.template.GetObjects(
				NewConfigMapSelector(
					func(d *corev1.ConfigMap) bool {
						return d.ObjectMeta.Name == name
					},
				),
			)
			if err != nil {
				return nil
			}
			return unsafeConvertConfigMaps(objects)
		},
	)

	configMaps := result.([]*corev1.ConfigMap)

	if len(configMaps) == 0 {
		return nil
	}
	return configMaps[0]
}

func (q *cachingQuery) ConfigMapsByLabels(labels map[string]string) []*corev1.ConfigMap {
	key := q.cacheKey(anything, gvkConfigMap, labels)
	result := q.runQuery(key,
		func() interface{} {
			objects, err := q.template.GetObjects(
				NewConfigMapSelector(
					func(d *corev1.ConfigMap) bool {
						return matchLabels(d.ObjectMeta.Labels, labels)
					},
				),
			)
			if err != nil {
				return nil
			}
			return unsafeConvertConfigMaps(objects)
		},
	)
	return result.([]*corev1.ConfigMap)
}

func (q *cachingQuery) ConfigMapByComponent(component string) *corev1.ConfigMap {
	configMaps := q.ConfigMapsByLabels(map[string]string{
		"app": component,
	})

	if len(configMaps) == 0 {
		return nil
	}
	return configMaps[0]
}

func (q *cachingQuery) JobByName(name string) *batchv1.Job {
	key := q.cacheKey(name, gvkJob, nil)
	result := q.runQuery(key,
		func() interface{} {
			objects, err := q.template.GetObjects(
				NewJobSelector(
					func(d *batchv1.Job) bool {
						return d.ObjectMeta.Name == name
					},
				),
			)
			if err != nil {
				return nil
			}
			return unsafeConvertJobs(objects)
		},
	)

	jobs := result.([]*batchv1.Job)

	if len(jobs) == 0 {
		return nil
	}
	return jobs[0]
}

func (q *cachingQuery) JobsByLabels(labels map[string]string) []*batchv1.Job {
	key := q.cacheKey(anything, gvkJob, labels)
	result := q.runQuery(key,
		func() interface{} {
			objects, err := q.template.GetObjects(
				NewJobSelector(
					func(d *batchv1.Job) bool {
						return matchLabels(d.ObjectMeta.Labels, labels)
					},
				),
			)
			if err != nil {
				return nil
			}
			return unsafeConvertJobs(objects)
		},
	)
	return result.([]*batchv1.Job)
}

func (q *cachingQuery) JobByComponent(component string) *batchv1.Job {
	jobs := q.JobsByLabels(map[string]string{
		"app": component,
	})

	if len(jobs) == 0 {
		return nil
	}
	return jobs[0]
}

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
		"app": component,
	})
	if len(deployments) == 0 {
		return nil
	}
	return deployments[0]
}

func (q *cachingQuery) StatefulSetByName(name string) *appsv1.StatefulSet {
	key := q.cacheKey(name, gvkStatefulSet, nil)
	result := q.runQuery(key,
		func() interface{} {
			objects, err := q.template.GetObjects(
				NewStatefulSetSelector(
					func(d *appsv1.StatefulSet) bool {
						return d.ObjectMeta.Name == name
					},
				),
			)
			if err != nil {
				return nil
			}
			return unsafeConvertStatefulSets(objects)
		},
	)

	statefulSets := result.([]*appsv1.StatefulSet)

	if len(statefulSets) == 0 {
		return nil
	}
	return statefulSets[0]
}

func (q *cachingQuery) StatefulSetsByLabels(labels map[string]string) []*appsv1.StatefulSet {
	key := q.cacheKey(anything, gvkStatefulSet, labels)
	result := q.runQuery(key,
		func() interface{} {
			objects, err := q.template.GetObjects(
				NewStatefulSetSelector(
					func(d *appsv1.StatefulSet) bool {
						return matchLabels(d.ObjectMeta.Labels, labels)
					},
				),
			)
			if err != nil {
				return nil
			}
			return unsafeConvertStatefulSets(objects)
		},
	)
	return result.([]*appsv1.StatefulSet)
}

func (q *cachingQuery) StatefulSetByComponent(component string) *appsv1.StatefulSet {
	statefulSets := q.StatefulSetsByLabels(map[string]string{
		"app": component,
	})
	if len(statefulSets) == 0 {
		return nil
	}
	return statefulSets[0]
}

func (q *cachingQuery) ServiceByName(name string) *corev1.Service {
	key := q.cacheKey(name, gvkService, nil)
	result := q.runQuery(key,
		func() interface{} {
			objects, err := q.template.GetObjects(
				NewServiceSelector(
					func(d *corev1.Service) bool {
						return d.ObjectMeta.Name == name
					},
				),
			)
			if err != nil {
				return nil
			}
			return unsafeConvertServices(objects)
		},
	)

	services := result.([]*corev1.Service)

	if len(services) == 0 {
		return nil
	}
	return services[0]
}

func (q *cachingQuery) ServicesByLabels(labels map[string]string) []*corev1.Service {
	key := q.cacheKey(anything, gvkService, labels)
	result := q.runQuery(key,
		func() interface{} {
			objects, err := q.template.GetObjects(
				NewServiceSelector(
					func(d *corev1.Service) bool {
						return matchLabels(d.ObjectMeta.Labels, labels)
					},
				),
			)
			if err != nil {
				return nil
			}
			return unsafeConvertServices(objects)
		},
	)
	return result.([]*corev1.Service)
}

func (q *cachingQuery) ServiceByComponent(component string) *corev1.Service {
	services := q.ServicesByLabels(map[string]string{
		"app": component,
	})
	if len(services) == 0 {
		return nil
	}
	return services[0]
}

func (q *cachingQuery) Reset() {
	q.clearCache()
}

func matchLabels(oLabels, qLabels map[string]string) bool {
	for k, v := range qLabels {
		if w, ok := oLabels[k]; !ok || v != w {
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

func unsafeConvertConfigMaps(objects []runtime.Object) []*corev1.ConfigMap {
	configMaps := make([]*corev1.ConfigMap, len(objects))
	for i, o := range objects {
		configMaps[i] = o.(*corev1.ConfigMap)
	}
	return configMaps
}

func unsafeConvertJobs(objects []runtime.Object) []*batchv1.Job {
	jobs := make([]*batchv1.Job, len(objects))
	for i, o := range objects {
		jobs[i] = o.(*batchv1.Job)
	}
	return jobs
}

func unsafeConvertSecrets(objects []runtime.Object) []*corev1.Secret {
	secrets := make([]*corev1.Secret, len(objects))
	for i, o := range objects {
		secrets[i] = o.(*corev1.Secret)
	}
	return secrets
}

func unsafeConvertDeployments(objects []runtime.Object) []*appsv1.Deployment {
	deployments := make([]*appsv1.Deployment, len(objects))
	for i, o := range objects {
		deployments[i] = o.(*appsv1.Deployment)
	}
	return deployments
}

func unsafeConvertStatefulSets(objects []runtime.Object) []*appsv1.StatefulSet {
	statefulSets := make([]*appsv1.StatefulSet, len(objects))
	for i, o := range objects {
		statefulSets[i] = o.(*appsv1.StatefulSet)
	}
	return statefulSets
}

func unsafeConvertServices(objects []runtime.Object) []*corev1.Service {
	services := make([]*corev1.Service, len(objects))
	for i, o := range objects {
		services[i] = o.(*corev1.Service)
	}
	return services
}

// CacheBackdoor is used by test cases.
func CacheBackdoor(q Query) *map[string]interface{} {
	if cq, ok := (q).(*cachingQuery); ok {
		return &cq.cache
	}
	return nil
}
