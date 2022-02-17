package gitlab

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
)

const (
	GitLabMailroomEnabled  = "gitlab.mailroom.enabled"
	IncomingEmailEnabled   = "global.appConfig.incomingEmail.enabled"
	IncomingEmailSecret    = "global.appConfig.incomingEmail.password.secret" //nolint:golint,gosec
	mailroomEnabledDefault = true
	incomingEmailDefault   = false
)

// MailroomEnabled returns `true` if enabled and `false` if not.
func MailroomEnabled(adapter CustomResourceAdapter) bool {
	mrEnabled := adapter.Values().GetBool(GitLabMailroomEnabled, mailroomEnabledDefault)
	imEnabled := adapter.Values().GetBool(IncomingEmailEnabled, incomingEmailDefault)
	emSecret := adapter.Values().GetString(IncomingEmailSecret, "")

	return mrEnabled && imEnabled && emSecret != ""
}

// MailroomDeployment returns the Deployment for the Mailroom component.
func MailroomDeployment(template helm.Template) client.Object {
	return template.Query().ObjectByKindAndComponent(DeploymentKind, MailroomComponentName)
}

// MailroomConfigMapsreturns the ConfigMaps for the Mailroom component.
func MailroomConfigMap(adapter CustomResourceAdapter, template helm.Template) client.Object {
	return template.Query().ObjectByKindAndName(ConfigMapKind,
		fmt.Sprintf("%s-%s", adapter.ReleaseName(), MailroomComponentName))
}
