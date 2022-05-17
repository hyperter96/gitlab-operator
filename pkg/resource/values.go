package resource

import (
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/support"
)

// ValueProvider offers a mechanism to adapt a Custom Resource and bind its
// specification to an arbitrary tree-like data structure. This structure can
// be consumed by any other component in the framework and in particular is
// useful for rendering templates, including Helm Charts.
type ValueProvider interface {
	// Values returns a mapping of a Custom Resource specification.
	//
	// You should not make any assumption about how the mapping is done. An
	// implementation may choose to do it eagerly or lazily. However, a mapping
	// must be one-way, meaning that the changes that are applied to it must not
	// propagate to the underlying resource.
	//
	// The provider does not watch the resource. It is the responsibility of the
	// caller to notify the provider upon changes. The provider may be able to
	// re-map the changed resources but it is not obligated to support it.
	Values() support.Values

	// Hash returns a unique hash for of the underlying Custom Resource.
	//
	// You should not make any assumption about when the hash calculated. An
	// implementation may choose to do it eagerly or lazily. However it must
	// guarantee that, after the completion of mapping, at any point in time
	// it returns the same value that does not change per mapping.
	Hash() string
}
