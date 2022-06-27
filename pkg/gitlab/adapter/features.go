package adapter

import (
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/gitlab"
)

// GitLabFeatures represents the features that are specified in the underlying
// GitLab resource. Note that these features are the desired status of an
// instance and does not necessarily mean that the instance exists or is in the
// desired state.
type GitLabFeatures interface {
	// WantsFeature queries this GitLab resource for the specified feature.
	// Returns true if the instance has the feature in its specification.
	//
	// Note that this function only checks the specification of the GitLab
	// resource and does not verify the state of the GitLab instance.
	WantsFeature(check gitlab.FeatureCheck) bool

	// WantsComponent is a shorthand for checking if a specific GitLab component
	// is enabled in the specification of this GitLab resource.
	//
	// Note that this function only checks the specification of the GitLab
	// resource and does not verify the state of the GitLab instance.
	WantsComponent(component gitlab.Component) bool
}
