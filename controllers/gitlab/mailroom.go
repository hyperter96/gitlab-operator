package gitlab

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2beta1 "k8s.io/api/autoscaling/v2beta1"
	corev1 "k8s.io/api/core/v1"
	networkpolicyv1 "k8s.io/api/networking/v1"
)

const (
	globalMailroomEnabled  = "global.mailroom.enabled"
	mailroomEnabledDefault = false
)

// MailroomEnabled returns `true` if enabled and `false` if not.
func MailroomEnabled(adapter CustomResourceAdapter) bool {
	enabled, _ := GetBoolValue(adapter.Values(), globalMailroomEnabled, mailroomEnabledDefault)
	return enabled
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

	exporterCfgMap := template.Query().ConfigMapByName(
		fmt.Sprintf("%s-%s", adapter.ReleaseName(), MailroomComponentName))

	result := []*corev1.ConfigMap{exporterCfgMap}

	return result[0]
}

// MailroomHPA returns the HPA for the Mailroom component.
func MailroomHPA(adapter CustomResourceAdapter) *autoscalingv2beta1.HorizontalPodAutoscaler {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	hpa := template.Query().HPAByName(
		fmt.Sprintf("%s-%s", adapter.ReleaseName(), MailroomComponentName))

	result := []*autoscalingv2beta1.HorizontalPodAutoscaler{hpa}

	return result[0]
}

// MailroomNetworkPolicy returns the NetworkPolicy for the Mailroom component.
func MailroomNetworkPolicy(adapter CustomResourceAdapter) []*networkpolicyv1.NetworkPolicy {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	policy := template.Query().NetworkPolicyByName(
		fmt.Sprintf("%s-%s", adapter.ReleaseName(), MailroomComponentName))

	result := []*networkpolicyv1.NetworkPolicy{policy}

	return result
}

// MailroomServiceAccount returns the ServiceAccount for the Mailroom component.
func MailroomServiceAccount(adapter CustomResourceAdapter) *corev1.ServiceAccount {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	account := template.Query().ServiceAccountByName(
		fmt.Sprintf("%s-%s", adapter.ReleaseName(), MailroomComponentName))

	result := []*corev1.ServiceAccount{account}

	return result[0]
}
