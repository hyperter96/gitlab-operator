package gitlab

import (
	"sigs.k8s.io/controller-runtime/pkg/client"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
)

func KasConfigMap(template helm.Template) client.Object {
	return template.Query().ObjectByKindAndComponent(ConfigMapKind, KasComponentName)
}

func KasDeployment(template helm.Template) client.Object {
	return template.Query().ObjectByKindAndComponent(DeploymentKind, KasComponentName)
}

func KasIngress(template helm.Template) client.Object {
	return template.Query().ObjectByKindAndComponent(IngressKind, KasComponentName)
}

func KasService(template helm.Template) client.Object {
	return template.Query().ObjectByKindAndComponent(ServiceKind, KasComponentName)
}
