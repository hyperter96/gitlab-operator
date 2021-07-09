package gitlab

import (
	"fmt"
)

const (
	// GitLabShellComponentName is the common name of GitLab Shell.
	GitLabShellComponentName = "gitlab-shell"

	// TaskRunnerComponentName is the common name of GitLab Task Runner.
	TaskRunnerComponentName = "task-runner"

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

	// PostgresComponentName is the common name of PostgreSQL.
	PostgresComponentName = "postgresql"

	// NGINXComponentName is the common name of NGINX Ingress.
	NGINXComponentName = "nginx-ingress"
)

func updateCommonLabels(releaseName, componentName string, labels map[string]string) {
	labels["app.kubernetes.io/name"] = releaseName
	labels["app.kubernetes.io/instance"] = fmt.Sprintf("%s-%s", releaseName, componentName)
	labels["app.kubernetes.io/component"] = componentName
	labels["app.kubernetes.io/part-of"] = "gitlab"
	labels["app.kubernetes.io/managed-by"] = "gitlab-operator"
}
