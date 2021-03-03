package gitlab

import (
	"fmt"
	"hash/fnv"
	"strings"

	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/helm"
)

// CustomResourceAdapter is a wrapper for GitLab Custom Resource. It provides a convenient interface
// to interact with the GitLab instances and guards the controller from its structural changes.
//
// This adapter is immutable and will not update itself after initialization. Therefore, it must be
// created when GitLab Custom Resource changes, e.g. in reconcile loop.
type CustomResourceAdapter interface {
	// Resource returns the reference to the underlaying Custom Resource.
	Resource() *gitlabv1beta1.GitLab

	// Hash generates a hash based on the key parts of a GitLab Custom Resource. The hash can be used
	// to identify changes to the underlaying resource. For example this is useful when rendering a
	// Helm template.
	Hash() string

	// Reference returns a fully qualified name of the associated GitLab Custom Resource. As opposed
	// to Hash this value does not change.
	Reference() string

	// Namespace returns the namespace in which the GitLab instance must be deployed. When Operator
	// is scoped to
	// a namespace this must be equal to the namespace of the Operator.
	Namespace() string

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
	result := &populatingAdapter{
		resource: gitlab,
		values:   helm.EmptyValues(),
	}
	result.populateValues()
	result.hashValues()
	return result
}

type populatingAdapter struct {
	resource  *gitlabv1beta1.GitLab
	values    helm.Values
	hash      string
	reference string
}

func (a *populatingAdapter) Resource() *gitlabv1beta1.GitLab {
	return a.resource
}

func (a *populatingAdapter) Hash() string {
	return a.hash
}

func (a *populatingAdapter) Reference() string {
	return a.reference
}

func (a *populatingAdapter) Namespace() string {
	return a.resource.Namespace
}

func (a *populatingAdapter) GitLabVersion() string {
	return a.resource.Spec.Release
}

func (a *populatingAdapter) ChartVersion() string {
	// Warning: This is a heuristic and may not work all the time.
	s := strings.Split(a.resource.Labels["chart"], "-")
	if len(s) < 2 {
		return ""
	}
	return s[len(s)-1]
}

func (a *populatingAdapter) ReleaseName() string {
	return a.resource.Name
}

func (a *populatingAdapter) Values() helm.Values {
	return a.values
}

func (a *populatingAdapter) populateValues() {
	a.reference = fmt.Sprintf("%s.%s", a.resource.Name, a.resource.Namespace)

	// Use auto-generated self-signed wildcard certificate
	a.values.AddValue("certmanager.install", "false")
	a.values.AddValue("global.ingress.configureCertmanager", "false")

	// Skip GitLab Runner
	a.values.AddValue("gitlab-runner.install", "false")

	// Set the default ImagePullPolicy
	a.values.AddValue("global.imagePullPolicy", "IfNotPresent")

	// Use NodePort Service type for GitLab Shell
	a.values.AddValue("gitlab.gitlab-shell.service.type", "NodePort")

	// Use manager ServiceAccount and local user for shared secrets
	a.values.AddValue("shared-secrets.serviceAccount.create", "false")
	a.values.AddValue("shared-secrets.serviceAccount.name", ManagerServiceAccount)
	a.values.AddValue("shared-secrets.securityContext.runAsUser", "")
	a.values.AddValue("shared-secrets.securityContext.fsGroup", "")
}

func (a *populatingAdapter) hashValues() {
	hasher := fnv.New64()
	valuesToHash := []([]byte){
		[]byte(a.Namespace()),
		[]byte(a.ReleaseName()),
		[]byte(a.ChartVersion()),

		// TODO: Marshal required values
	}
	valuesHashed := 0

	for _, v := range valuesToHash {
		_, err := hasher.Write(v)

		if err == nil {
			valuesHashed++
		}
	}

	if valuesHashed == 0 {
		// This is here to cover all the bases. Otherwise it should never happen.
		a.hash = fmt.Sprintf("%s/%s", a.ChartVersion(), a.GitLabVersion())
	}

	a.hash = fmt.Sprintf("%x", hasher.Sum64())
}
