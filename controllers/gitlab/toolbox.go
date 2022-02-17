package gitlab

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
)

const (
	gitlabToolboxEnabled                   = "gitlab.%s.enabled"
	toolboxEnabledDefault                  = true
	gitlabToolboxCronJobEnabled            = "gitlab.%s.backups.cron.enabled"
	toolboxCronJobEnabledDefault           = false
	gitlabToolboxCronJobPersistenceEnabled = "gitlab.%s.backups.cron.persistence.enabled"
	gitlabToolboxCronJobPersistenceDefault = false
)

// ToolboxEnabled returns `true` if Toolbox is enabled, and `false` if not.
func ToolboxEnabled(adapter CustomResourceAdapter) bool {
	key := fmt.Sprintf(gitlabToolboxEnabled, ToolboxComponentName(adapter.ChartVersion()))

	return adapter.Values().GetBool(key, toolboxEnabledDefault)
}

// ToolboxCronJobEnabled returns `true` if Toolbox CronJob is enabled, and `false` if not.
func ToolboxCronJobEnabled(adapter CustomResourceAdapter) bool {
	key := fmt.Sprintf(gitlabToolboxCronJobEnabled, ToolboxComponentName(adapter.ChartVersion()))

	return adapter.Values().GetBool(key, toolboxCronJobEnabledDefault)
}

// ToolboxCronJobPersistenceEnabled returns `true` if Toolbox CronJob persistence is enabled, and `false` if not.
func ToolboxCronJobPersistenceEnabled(adapter CustomResourceAdapter) bool {
	key := fmt.Sprintf(gitlabToolboxCronJobPersistenceEnabled, ToolboxComponentName(adapter.ChartVersion()))

	return adapter.Values().GetBool(key, gitlabToolboxCronJobPersistenceDefault)
}

// ToolboxDeployment returns the Deployment of the Toolbox component.
func ToolboxDeployment(adapter CustomResourceAdapter, template helm.Template) client.Object {
	return template.Query().ObjectByKindAndComponent(DeploymentKind, ToolboxComponentName(adapter.ChartVersion()))
}

// ToolboxConfigMap returns the ConfigMaps of the Toolbox component.
func ToolboxConfigMap(adapter CustomResourceAdapter, template helm.Template) client.Object {
	return template.Query().ObjectByKindAndName(ConfigMapKind,
		fmt.Sprintf("%s-%s", adapter.ReleaseName(), ToolboxComponentName(adapter.ChartVersion())))
}

// ToolboxCronJob returns the CronJob of the Toolbox component.
func ToolboxCronJob(adapter CustomResourceAdapter, template helm.Template) client.Object {
	return template.Query().ObjectByKindAndName(CronJobKind,
		fmt.Sprintf("%s-%s-backup", adapter.ReleaseName(), ToolboxComponentName(adapter.ChartVersion())))
}

// ToolboxPersistentVolumeClaim returns the PersistentVolumeClaim of the Toolbox component.
func ToolboxCronJobPersistentVolumeClaim(adapter CustomResourceAdapter, template helm.Template) client.Object {
	return template.Query().ObjectByKindAndName(PersistentVolumeClaimKind,
		fmt.Sprintf("%s-%s-backup-tmp", adapter.ReleaseName(), ToolboxComponentName(adapter.ChartVersion())))
}
