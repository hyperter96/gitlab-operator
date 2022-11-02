package objects

import (
	"k8s.io/apimachinery/pkg/runtime/schema"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Selector is a function that receives a Kubernetes resource and checks whether
// it satisfies a condition, including type requirements. It returns true when
// the object meets the expectations.
type Selector = func(client.Object) bool

// ByKind returns a selector that matches a Kubernetes resource kind with the
// kind that is specified.
//
// It accepts the common style of string that is used by `apimachinery` which
// can be `Kind`, `Kind.group.tld` or `Kind.version.group.tld`.
func ByKind(kind string) Selector {
	return func(o client.Object) bool {
		result := false
		qGVK, qGK := schema.ParseKindArg(kind)
		oGVK := o.GetObjectKind().GroupVersionKind()

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
}

// ByName returns a selector that matches a Kubernetes resource name with the
// name that is specified.
//
// It does not match resource namespace or kind.
func ByName(name string) Selector {
	return func(o client.Object) bool {
		return o.GetName() == name
	}
}

// ByNamespace returns a selector that matches a Kubernetes resource namespace
// with the namespace that is specified.
//
// It does not match resource name or kind.
func ByNamespace(namespace string) Selector {
	return func(o client.Object) bool {
		return o.GetNamespace() == namespace
	}
}

// ByLabels returns a selector that matches a Kubernetes resource labels
// with the labels that are specified.
//
// It does not match resource name or kind and only checks if the resource
// labels are inclusive of the specified labels, i.e. the specified labels are
// a subset of the resource labels. It is similar to the label selector of
// option of `kubectl` command-line.
func ByLabels(labels map[string]string) Selector {
	return func(o client.Object) bool {
		oLabels := o.GetLabels()

		for k, v := range labels {
			if w, ok := oLabels[k]; !ok || v != w {
				return false
			}
		}

		return true
	}
}

// ByComponent returns a selector that matches a Kubernetes resource labels
// with the component that is specified.
//
// It does not match resource name or kind and only checks if the resource
// has `app` or `app.kubernetes.io/component` labels with the specified
// component.
func ByComponent(component string) Selector {
	return func(o client.Object) bool {
		oLabels := o.GetLabels()

		return oLabels["app"] == component ||
			oLabels["app.kubernetes.io/component"] == component
	}
}

// All combines the provided object selectors and succeeds when all of them
// return true.
func All(selectors ...Selector) Selector {
	return func(o client.Object) bool {
		for _, selector := range selectors {
			if !selector(o) {
				return false
			}
		}

		return true
	}
}

// Any combines the provided object selectors and succeeds when any of them
// returns true.
func Any(selectors ...Selector) Selector {
	return func(o client.Object) bool {
		for _, selector := range selectors {
			if selector(o) {
				return true
			}
		}

		return false
	}
}

// None combines the provided object selectors and succeeds when none of them
// return true.
func None(selectors ...Selector) Selector {
	return func(o client.Object) bool {
		for _, selector := range selectors {
			if selector(o) {
				return false
			}
		}

		return true
	}
}

// Negate changes the specified selector and negates its result.
func Negate(predicate Selector) Selector {
	return func(o client.Object) bool {
		return !predicate(o)
	}
}
