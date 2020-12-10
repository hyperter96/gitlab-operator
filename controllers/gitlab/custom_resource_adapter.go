package gitlab

import (
	"fmt"
	"hash/fnv"
	"strings"

	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/helm"
)

// CustomResourceAdapter is a wrapper for GitLab Custom Resource. It provides a convenient interface
// to interact with the GitLab instances and guards the controller from its structural changes.
//
// This adapter is immutable and will not update itself after initialization. Therefore, it must be
// created when GitLab Custom Resource changes, e.g. in reconcile loop.
type CustomResourceAdapter interface {
	// Hash generates a hash based on the key parts of a GitLab Custom Resource. The hash can be used
	// to identify changes to the underlaying resource. For example this is useful when rendering a
	// Helm template.
	Hash() string

	// Namespace returns the namespace in which the GitLab instance must be deployed. When Operator
	// is scoped to
	// a namespace this must be equal to the namespace of the Operator.
	Nampespace() string

	// ChartVersion returns the version of GitLab chart that must be used to deploy this GitLab
	// instance.
	ChartVersion() string

	// GitLabVersion returns the version of GitLab. This is generally derived from the GitLab chart.
	GitLabVersion() string

	// ReleaseName returns the name of the GitLab instance that must be deployed. This will be used
	// as a qualifier to distinguish between multiple GitLab instances in a namespace.
	ReleaseName() string

	// Values returns the set of values that will be used the render GitLab chart.
	Values() helm.Values
}

// NewCustomResourceAdapter returns a new adapter for the provided GitLab instance.
func NewCustomResourceAdapter(gitlab *gitlabv1beta1.GitLab) CustomResourceAdapter {
	result := &wrappingCustomResourceAdapter{
		gitlab: gitlab,
		values: helm.EmptyValues(),
	}
	result.populateValues()
	return result
}

type wrappingCustomResourceAdapter struct {
	gitlab *gitlabv1beta1.GitLab
	values helm.Values
}

func (w *wrappingCustomResourceAdapter) Hash() string {
	hasher := fnv.New64()
	valuesToHash := []([]byte){
		[]byte(w.ChartVersion()),
		[]byte(w.GitLabVersion()),
		// Marshal values
	}
	valuesHashed := 0

	for _, v := range valuesToHash {
		_, err := hasher.Write(v)

		if err == nil {
			valuesHashed++
		}
	}

	if valuesHashed == 0 {
		return fmt.Sprintf("%s/%s", w.ChartVersion(), w.GitLabVersion())
	}

	return fmt.Sprintf("%x", hasher.Sum64())
}

func (w *wrappingCustomResourceAdapter) Nampespace() string {
	return w.gitlab.Namespace
}

func (w *wrappingCustomResourceAdapter) GitLabVersion() string {
	return w.gitlab.Spec.Release
}

func (w *wrappingCustomResourceAdapter) ChartVersion() string {
	// Warning: This is a heuristic and may not work all the time.
	s := strings.Split(w.gitlab.Labels["chart"], "-")
	if len(s) < 2 {
		return ""
	}
	return s[len(s)-1]
}

func (w *wrappingCustomResourceAdapter) ReleaseName() string {
	return w.gitlab.Labels["release"]
}

func (w *wrappingCustomResourceAdapter) Values() helm.Values {
	return w.values
}

func (w *wrappingCustomResourceAdapter) populateValues() {
	// Read values for rendering Helm template from GitLab Custom Resource.
}
