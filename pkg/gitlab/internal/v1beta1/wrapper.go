package v1beta1

import (
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

/* CustomResourceWrapper */

func (w *Adapter) Name() types.NamespacedName {
	return types.NamespacedName{
		Name:      w.source.Name,
		Namespace: w.source.Namespace,
	}
}

func (w *Adapter) Origin() client.Object {
	return w.source
}
