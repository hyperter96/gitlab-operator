package gitlab

import (
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

func PrometheusDeployments(template helm.Template) []client.Object {
	return template.Query().ObjectsByKindAndLabels(DeploymentKind, map[string]string{
		"app": PrometheusComponentName,
	})
}

func PrometheusStatefulSets(template helm.Template) []client.Object {
	return template.Query().ObjectsByKindAndLabels(StatefulSetKind, map[string]string{
		"app": PrometheusComponentName,
	})
}

func PrometheusDaemonSets(template helm.Template) []client.Object {
	return template.Query().ObjectsByKindAndLabels(DaemonSetKind, map[string]string{
		"app": PrometheusComponentName,
	})
}

func PrometheusServices(template helm.Template) []client.Object {
	return template.Query().ObjectsByKindAndLabels(ServiceKind, map[string]string{
		"app": PrometheusComponentName,
	})
}

func PrometheusConfigMaps(template helm.Template) []client.Object {
	return template.Query().ObjectsByKindAndLabels(ConfigMapKind, map[string]string{
		"app": PrometheusComponentName,
	})
}

func PrometheusPersistentVolumeClaims(template helm.Template) []client.Object {
	return template.Query().ObjectsByKindAndLabels(PersistentVolumeClaimKind, map[string]string{
		"app": PrometheusComponentName,
	})
}

func PrometheusIngresses(template helm.Template) []client.Object {
	return template.Query().ObjectsByKindAndLabels(IngressKind, map[string]string{
		"app": PrometheusComponentName,
	})
}
