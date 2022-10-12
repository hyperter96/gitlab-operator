package gitlab

import (
	"strings"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/gitlab"
)

const (
	GlobalPraefectVirtualStorages              = "global.praefect.virtualStorages"
	GlobalPraefectReplaceInternalGitalyEnabled = "global.praefect.replaceInternalGitaly"
)

// PraefectReplaceInternalGitalyEnabled returns `true` if enabled and `false` if not.
func PraefectReplaceInternalGitalyEnabled(adapter gitlab.Adapter) bool {
	return adapter.Values().GetBool(GlobalPraefectReplaceInternalGitalyEnabled)
}

// GitalyPraefectConfigMap returns the Gitaly ConfigMap of Praefect component.
func GitalyPraefectConfigMap(template helm.Template) client.Object {
	results := template.Query().ObjectsByKindAndLabels(ConfigMapKind, map[string]string{"app": GitalyComponentName})

	for _, result := range results {
		if hasPraefectSuffix := strings.HasSuffix(result.GetName(), "-praefect"); hasPraefectSuffix {
			return result
		}
	}

	return nil
}

// GitalyPraefectServices returns the Gitaly Services of Praefect component.
func GitalyPraefectServices(template helm.Template) []client.Object {
	result := template.Query().ObjectsByKindAndLabels(ServiceKind, map[string]string{
		"app": GitalyComponentName,
	})

	var results []client.Object = nil

	for _, service := range result {
		if _, hasStorageLabel := service.GetLabels()["storage"]; hasStorageLabel {
			results = append(results, service)
		}
	}

	return results
}

// GitalyPraefectStatefulSets returns the Gitaly StatefulSets of Praefect component.
func GitalyPraefectStatefulSets(template helm.Template) []client.Object {
	result := template.Query().ObjectsByKindAndLabels(StatefulSetKind, map[string]string{
		"app": GitalyComponentName,
	})

	var results []client.Object = nil

	for _, statefulSet := range result {
		if _, hasStorageLabel := statefulSet.GetLabels()["storage"]; hasStorageLabel {
			results = append(results, statefulSet)
		}
	}

	return results
}
