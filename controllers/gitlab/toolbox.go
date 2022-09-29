package gitlab

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/gitlab"
)

// ToolboxDeployment returns the Deployment of the Toolbox component.
func ToolboxDeployment(adapter gitlab.Adapter, template helm.Template) client.Object {
	return template.Query().ObjectByKindAndComponent(DeploymentKind, ToolboxComponentName)
}

// ToolboxConfigMap returns the ConfigMaps of the Toolbox component.
func ToolboxConfigMap(adapter gitlab.Adapter, template helm.Template) client.Object {
	return template.Query().ObjectByKindAndName(ConfigMapKind,
		fmt.Sprintf("%s-%s", adapter.ReleaseName(), ToolboxComponentName))
}

// ToolboxCronJob returns the CronJob of the Toolbox component.
func ToolboxCronJob(adapter gitlab.Adapter, template helm.Template) client.Object {
	return template.Query().ObjectByKindAndName(CronJobKind,
		fmt.Sprintf("%s-%s-backup", adapter.ReleaseName(), ToolboxComponentName))
}

// ToolboxPersistentVolumeClaim returns the PersistentVolumeClaim of the Toolbox component.
func ToolboxCronJobPersistentVolumeClaim(adapter gitlab.Adapter, template helm.Template) client.Object {
	return template.Query().ObjectByKindAndName(PersistentVolumeClaimKind,
		fmt.Sprintf("%s-%s-backup-tmp", adapter.ReleaseName(), ToolboxComponentName))
}
