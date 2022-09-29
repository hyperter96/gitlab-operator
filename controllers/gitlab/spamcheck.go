package gitlab

import (
	"sigs.k8s.io/controller-runtime/pkg/client"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
)

func SpamcheckConfigMap(template helm.Template) client.Object {
	return template.Query().ObjectByKindAndComponent(ConfigMapKind, SpamcheckComponentName)
}

func SpamcheckDeployment(template helm.Template) client.Object {
	return template.Query().ObjectByKindAndComponent(DeploymentKind, SpamcheckComponentName)
}

func SpamcheckService(template helm.Template) client.Object {
	return template.Query().ObjectByKindAndComponent(ServiceKind, SpamcheckComponentName)
}
