package gitlab

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
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
func MailroomDeployment(adapter CustomResourceAdapter) *appsv1.Deployment {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	result := template.Query().DeploymentByComponent(MailroomComponentName)

	return result
}

// MailroomConfigMapsreturns the ConfigMaps for the Mailroom component.
func MailroomConfigMap(adapter CustomResourceAdapter) *corev1.ConfigMap {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	mailroomCfgMap := template.Query().ConfigMapByName(
		fmt.Sprintf("%s-%s", adapter.ReleaseName(), MailroomComponentName))

	return mailroomCfgMap
}

// MailroomNetworkPolicy returns the NetworkPolicy for the Mailroom component.
func MailroomNetworkPolicy(adapter CustomResourceAdapter) *networkingv1.NetworkPolicy {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	policy := template.Query().NetworkPolicyByName(
		fmt.Sprintf("%s-%s", adapter.ReleaseName(), MailroomComponentName))

	return policy
}
