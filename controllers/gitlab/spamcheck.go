package gitlab

import (
	"sigs.k8s.io/controller-runtime/pkg/client"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/gitlab"
)

const (
	globalSpamcheckEnabled = "global.spamcheck.enabled"
)

func SpamcheckEnabled(adapter gitlab.Adapter) bool {
	return adapter.Values().GetBool(globalSpamcheckEnabled)
}

func SpamcheckConfigMap(template helm.Template) client.Object {
	return template.Query().ObjectByKindAndComponent(ConfigMapKind, SpamcheckComponentName)
}

func SpamcheckDeployment(template helm.Template) client.Object {
	return template.Query().ObjectByKindAndComponent(DeploymentKind, SpamcheckComponentName)
}

func SpamcheckService(template helm.Template) client.Object {
	return template.Query().ObjectByKindAndComponent(ServiceKind, SpamcheckComponentName)
}
