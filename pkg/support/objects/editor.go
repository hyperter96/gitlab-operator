package objects

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Editor is a function that receives a Kubernetes resource and changes it
// in-place. It is used by the Edit function of Collection type where it accepts
// multiple editors.
//
// It must return an error if an unhandled error occurs while editing. It can
// choose to return a type mismatch error, using `NewTypeMistmatchError`, to
// signal that it can not change the object due to type restrictions. This
// particular error is ignored by the Edit function.
//
// An editor should not assume the order of the execution in a chain of editors
// and ideally should be idempotent.
type Editor = func(client.Object) error

// SetNamespace returns an Editor that sets the namespace of a Kubernetes
// resource.
func SetNamespace(namespace string) Editor {
	return func(o client.Object) error {
		o.SetNamespace(namespace)

		return nil
	}
}

// SetAnnotations returns an Editor that sets the specified annotations of a
// Kubernetes resource.
//
// It does not replace all the annotations, instead it merges the specified
// annotations with the existing annotations of the object.
func SetAnnotations(annotations map[string]string) Editor {
	return func(o client.Object) error {
		oAnnotations := o.GetAnnotations()

		for k, v := range annotations {
			oAnnotations[k] = v
		}

		o.SetAnnotations(oAnnotations)

		return nil
	}
}

// NewTypeMismatchError creates an error that is used by Editor to signal that
// it can not change the object due to type restrictions.
func NewTypeMismatchError(expected, observed interface{}) error {
	return &typeMismatchError{
		expected: expected,
		observed: observed,
	}
}

// IsTypeMismatchError returns true when an error is a type mismatch error.
func IsTypeMismatchError(err error) bool {
	_, ok := err.(*typeMismatchError)
	return ok
}

type typeMismatchError struct {
	expected interface{}
	observed interface{}
}

func (e *typeMismatchError) Error() string {
	return fmt.Sprintf("expected %T, got %T", e.expected, e.observed)
}
