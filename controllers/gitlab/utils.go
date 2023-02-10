package gitlab

import (
	"fmt"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/gitlab"
)

const (
	// Known object kinds.
	ConfigMapKind               = "ConfigMap"
	CronJobKind                 = "CronJob"
	DeploymentKind              = "Deployment"
	HorizontalPodAutoscalerKind = "HorizontalPodAutoscaler"
	IngressKind                 = "Ingress"
	JobKind                     = "Job"
	PersistentVolumeClaimKind   = "PersistentVolumeClaim"
	SecretKind                  = "Secret"
	ServiceKind                 = "Service"
	StatefulSetKind             = "StatefulSet"

	// GitlabComponentName is the com mon name of GitLab.
	GitLabComponentName = "gitlab"

	// GitLabShellComponentName is the common name of GitLab Shell.
	GitLabShellComponentName = "gitlab-shell"

	// MigrationsComponentName is the common name of Migrations.
	MigrationsComponentName = "migrations"

	// GitLabExporterComponentName is the common name of GitLab Exporter.
	GitLabExporterComponentName = "gitlab-exporter"

	// RegistryComponentName is the common name of the Registry.
	RegistryComponentName = "registry"

	// WebserviceComponentName is the common name of Webservice.
	WebserviceComponentName = "webservice"

	// SharedSecretsComponentName is the common name of Shared Secrets.
	SharedSecretsComponentName = "shared-secrets"

	// GitalyComponentName is the common name of Gitaly.
	GitalyComponentName = "gitaly"

	// SidekiqComponentName is the common name of Sidekiq.
	SidekiqComponentName = "sidekiq"

	// RedisComponentName is the common name of Redis.
	RedisComponentName = "redis"

	// DefaultPostgresComponentName is the default common name of PostgreSQL.
	DefaultPostgresComponentName = "postgresql"

	// NGINXComponentName is the common name of NGINX Ingress.
	NGINXComponentName = "nginx-ingress"

	// NGINXDefaultBackendComponentName is the common name of NGINX DefaultBackend.
	NGINXDefaultBackendComponentName = "defaultbackend"

	// PagesComponentName is the common name of GitLab Pages.
	PagesComponentName = "gitlab-pages"

	// PraefectComponentName is the common name of Praefect.
	PraefectComponentName = "praefect"

	// MailroomComponentName is the common name of Mailroom.
	MailroomComponentName = "mailroom"

	// KasComponentName is the common name of KAS.
	KasComponentName = "kas"

	// ToolboxComponentName is the common name of Toolbox.
	ToolboxComponentName = "toolbox"

	// MinioComponentName is the common name of MinIO.
	MinioComponentName = "minio"

	// SpamcheckComponentName is the common name of Spamcheck.
	SpamcheckComponentName = "spamcheck"
)

// RedisSubqueues is the array of possible Redis subqueues.
func RedisSubqueues() [5]string {
	return [5]string{"cache", "sharedState", "queues", "actioncable", "traceChunks"}
}

// PostgresComponentName provides the name of the PostgreSQL component taking name overrides into account.
// Attention: This component name is not part of all PostgreSQL ressource names.
func PostgresComponentName(adapter gitlab.Adapter) string {
	if nameOverride := adapter.Values().GetString("postgresql.nameOverride", ""); nameOverride != "" {
		return nameOverride
	}

	return DefaultPostgresComponentName
}

func updateCommonLabels(releaseName, componentName string, labels map[string]string) {
	labels["app.kubernetes.io/name"] = releaseName
	labels["app.kubernetes.io/instance"] = fmt.Sprintf("%s-%s", releaseName, componentName)
	labels["app.kubernetes.io/component"] = componentName
	labels["app.kubernetes.io/part-of"] = "gitlab"
	labels["app.kubernetes.io/managed-by"] = "gitlab-operator"
}
