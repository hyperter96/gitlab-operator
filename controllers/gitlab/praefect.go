package gitlab

import (
	"sigs.k8s.io/controller-runtime/pkg/client"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
)

// PraefectStatefulSet returns the StatefulSet of Praefect component.
func PraefectStatefulSet(template helm.Template) client.Object {
	return template.Query().ObjectByKindAndComponent(StatefulSetKind, PraefectComponentName)
}

// PraefectConfigMap returns the ConfigMap of Praefect component.
func PraefectConfigMap(template helm.Template) client.Object {
	return template.Query().ObjectByKindAndComponent(ConfigMapKind, PraefectComponentName)
}

// PraefectService returns the Service of GitLab Praefect component.
func PraefectService(template helm.Template) client.Object {
	return template.Query().ObjectByKindAndComponent(ServiceKind, PraefectComponentName)
}
