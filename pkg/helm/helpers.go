package helm

import (
	"github.com/pkg/errors"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

// AnySelector is an ObjectSelector that selects all objects.
var AnySelector = func(_ runtime.Object) bool {
	return true
}

// NoneSelector is an ObjectSelector that selects no object.
var NoneSelector = func(_ runtime.Object) bool {
	return false
}

// ServiceSelector is an ObjectSelector that selects Service objects.
var ServiceSelector = func(o runtime.Object) bool {
	_, ok := o.(*core.Service)
	return ok
}

// DeploymentSelector is an ObjectSelector that selects Deployment objects.
var DeploymentSelector = func(o runtime.Object) bool {
	_, ok := o.(*apps.Deployment)
	return ok
}

// StatefulSetSelector is an ObjectSelector that selects StatefulSet objects.
var StatefulSetSelector = func(o runtime.Object) bool {
	_, ok := o.(*apps.StatefulSet)
	return ok
}

// IngressSelector is an ObjectSelector that selects Ingress objects.
var IngressSelector = func(o runtime.Object) bool {
	_, ok := o.(*networking.Ingress)
	return ok
}

// NewServiceEditor builds a new ObjectEditor for Service objects with the provided delegate.
func NewServiceEditor(delegate func(*core.Service) error) ObjectEditor {
	return func(o *runtime.Object) error {
		service, ok := (*o).(*core.Service)
		if !ok {
			return errors.Errorf("expected %T, got %T", core.Service{}, *o)
		}
		return delegate(service)
	}
}

// NewDeploymentEditor builds a new ObjectEditor for Deployment objects with the provided delegate.
func NewDeploymentEditor(delegate func(*apps.Deployment) error) ObjectEditor {
	return func(o *runtime.Object) error {
		deployment, ok := (*o).(*apps.Deployment)
		if !ok {
			return errors.Errorf("expected %T, got %T", apps.Deployment{}, *o)
		}
		return delegate(deployment)
	}
}
