package resource

import (
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CustomResourceWrapper is a general purpose wrapper for Custom Resources. It
// provides a semantic interface to interact with resource instances and guards
// the controllers against the structural changes to Custom Resource Definition
// while satisfying certain behaviors from Custom Resources.
//
// A wrapper is immutable and does not update itself after initialization.
// Therefore, it must be created when the underlying resource changes, for
// example at the very beginning of the reconcile loop.
type CustomResourceWrapper interface {
	// Name is a convenient method that returns the fully qualified name of the
	// underlying Custom Resource.
	Name() types.NamespacedName

	// Origin returns the reference to the underlying Custom Resource.
	Origin() client.Object
}
