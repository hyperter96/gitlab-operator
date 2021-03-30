package internal

import (
	"reflect"

	"github.com/imdario/mergo"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// IsDeploymentChanged compares two deployments
// and returns a merged deployment object and true if they are different
func IsDeploymentChanged(old, new *appsv1.Deployment) (*appsv1.Deployment, bool) {

	if err := mergo.Merge(new, *old); err != nil {
		return old, false
	}

	if !reflect.DeepEqual(new.Spec.Template.Spec.InitContainers,
		old.Spec.Template.Spec.InitContainers) {
		return new, true
	}

	if !reflect.DeepEqual(new.Spec.Template.Spec.Containers,
		old.Spec.Template.Spec.Containers) {
		return new, true
	}

	if !reflect.DeepEqual(new.Spec.Template.Spec.Volumes,
		old.Spec.Template.Spec.Volumes) {
		return new, true
	}

	return new, !reflect.DeepEqual(old.Spec.Template.Annotations, new.Spec.Template.Annotations) ||
		!reflect.DeepEqual(old.ObjectMeta.Labels, new.ObjectMeta.Labels)
}

// IsConfigMapChanged returns an  updated configmap object
// and true if the configmap has changed
func IsConfigMapChanged(old, new *corev1.ConfigMap) (*corev1.ConfigMap, bool) {
	if err := mergo.Merge(new, *old); err != nil {
		return old, false
	}

	checksum, ok := old.ObjectMeta.Annotations["checksum"]
	if ok {
		if checksum == new.ObjectMeta.Annotations["checksum"] {
			return old, false
		}
	}

	if reflect.DeepEqual(old, new) {
		return old, false
	}

	return new, true
}
