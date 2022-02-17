package gitlab

import (
	"sigs.k8s.io/controller-runtime/pkg/client"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
)

const (
	GlobalGitalyEnabled  = "global.gitaly.enabled"
	gitalyEnabledDefault = true
)

// GitalyEnabled returns `true` if enabled and `false` if not.
func GitalyEnabled(adapter CustomResourceAdapter) bool {
	return adapter.Values().GetBool(GlobalGitalyEnabled, gitalyEnabledDefault)
}

// GitalyStatefulSet returns the StatefulSet of Gitaly component.
func GitalyStatefulSet(template helm.Template) client.Object {
	return template.Query().ObjectByKindAndComponent(StatefulSetKind, GitalyComponentName)
}

// GitalyConfigMap returns the ConfigMap of Gitaly component.
func GitalyConfigMap(template helm.Template) client.Object {
	return template.Query().ObjectByKindAndComponent(ConfigMapKind, GitalyComponentName)
}

// GitalyService returns the Service of GitLab Shell component.
func GitalyService(template helm.Template) client.Object {
	return template.Query().ObjectByKindAndComponent(ServiceKind, GitalyComponentName)
}
