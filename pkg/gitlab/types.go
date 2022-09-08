package gitlab

import (
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/support"
)

// ConditionType is an alias type for representing the type of a GitLab status
// condition.
type ConditionType string

// Component is an alias type for representing an individual GitLab component.
type Component string

// Components is a type for grouping and addressing a collection of GitLab
// components.
type Components []Component

// FeatureCheck is a callback for assessing the availability of a GitLab feature
// based on the values of specification of a GitLab resource.
type FeatureCheck func(values support.Values) bool

// Name returns the name of the condition.
func (c ConditionType) Name() string {
	return string(c)
}

// Name returns the name of the component.
func (c Component) Name() string {
	return string(c)
}

// Names returns the name of the components in the collection in the same order.
func (c Components) Names() []string {
	result := make([]string, len(c))
	for i := 0; i < len(c); i++ {
		result[i] = c[i].Name()
	}

	return result
}

// Contains checks wheather the collection contains the specified component.
func (c Components) Contains(component Component) bool {
	for _, i := range c {
		if i == component {
			return true
		}
	}

	return false
}
