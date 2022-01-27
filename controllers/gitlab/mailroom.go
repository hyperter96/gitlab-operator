package gitlab

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"
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
func MailroomDeployment(adapter CustomResourceAdapter) client.Object {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	result := template.Query().ObjectByKindAndComponent(DeploymentKind, MailroomComponentName)

	return result
}

// MailroomConfigMapsreturns the ConfigMaps for the Mailroom component.
func MailroomConfigMap(adapter CustomResourceAdapter) client.Object {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	mailroomCfgMap := template.Query().ObjectByKindAndName(ConfigMapKind,
		fmt.Sprintf("%s-%s", adapter.ReleaseName(), MailroomComponentName))

	return mailroomCfgMap
}
