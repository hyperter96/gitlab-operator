package helm

import (
	"fmt"
	"sync"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Query provides access methods to query Helm templates.
type Query interface {
	// Template returns the attached template that this interface queries.
	Template() Template

	// Reset clears the query cache when applicable.
	Reset()

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

	// DeploymentByName returns the Deployment with the specified name.
	DeploymentByName(name string) *appsv1.Deployment

	// DeploymentsByLabels lists all Deployments that match the labels.
	DeploymentsByLabels(labels map[string]string) []*appsv1.Deployment

	// DeploymentByComponent returns the Deployment for a specific component.
	DeploymentByComponent(component string) *appsv1.Deployment

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

	// ServiceByName returns the Service with the specified name.
	ServiceByName(name string) *corev1.Service

	// ServicesByLabels lists all Services that match the labels.
	ServicesByLabels(labels map[string]string) []*corev1.Service

	// ServiceByComponent returns the Service for a specific component.
	ServiceByComponent(component string) *corev1.Service

	// StatefulSetByName returns the StatefulSet with the specified name.
	StatefulSetByName(name string) *appsv1.StatefulSet

	// StatefulSetsByLabels lists all StatefulSets that match the labels.
	StatefulSetsByLabels(labels map[string]string) []*appsv1.StatefulSet

	// StatefulSetByComponent returns the StatefulSet for a specific component.
	StatefulSetByComponent(component string) *appsv1.StatefulSet

	// IngressesByLabels lists all Ingresses that match the labels.
	IngressesByLabels(labels map[string]string) []*extensionsv1beta1.Ingress

	// IngressByComponent returns the INgress for a specific component.
	IngressByComponent(component string) *extensionsv1beta1.Ingress
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
	anything = "*"

	gvkDeployment              = "Deployment.v1.apps"
	gvkStatefulSet             = "StatefulSet.v1.apps"
	gvkJob                     = "Job.v1.batch"
	gvkConfigMap               = "ConfigMap.v1.core"
	gvkSecret                  = "Secret.v1.core"
	gvkService                 = "Service.v1.core"
	gvkIngress                 = "Ingress.v1beta1.extensions"
	gvkHorizontalPodAutoscaler = "HorizontalPodAutoscaler.v1.autoscaling"

	appLabel = "app"
)

func (q *cachingQuery) Template() Template {
	return q.template
}

func (q *cachingQuery) Reset() {
	q.clearCache()
}

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

func matchLabels(oLabels, qLabels map[string]string) bool {
	for k, v := range qLabels {
		if w, ok := oLabels[k]; !ok || v != w {
			return false
		}
	}

	return true
}

func cacheBackdoor(q Query) *map[string]interface{} {
	if cq, ok := (q).(*cachingQuery); ok {
		return &cq.cache
	}

	return nil
}
