package gitlab

import (
	"fmt"
	"strings"

	"github.com/Masterminds/semver"
)

const (
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

	// PostgresComponentName is the common name of PostgreSQL.
	PostgresComponentName = "postgresql"

	// NGINXComponentName is the common name of NGINX Ingress.
	NGINXComponentName = "nginx-ingress"

	// PagesComponentName is the common name of GitLab Pages.
	PagesComponentName = "gitlab-pages"

	// MailroomComponentName is the common name of Mailroom.
	MailroomComponentName = "mailroom"

	// KasComponentName is the common name of KAS.
	KasComponentName = "kas"
)

// RedisSubqueues is the array of possible Redis subqueues.
func RedisSubqueues() [5]string {
	return [5]string{"cache", "sharedState", "queues", "actioncable", "traceChunks"}
}

// ToolboxComponentName returns the component name for Toolbox depending on the Chart version.
// If the Chart version is >= 5.5.0, then it returns "toolbox".
// If the Chart version is < 5.5.0, then it returns "task-runner".
// When the list of supported Chart versions are all 5.5.0 or newer, this function
// can be removed and we can use a constant `ToolboxComponentName = "toolbox"`.
func ToolboxComponentName(chartVersion string) string {
	versionWithToolbox, _ := semver.NewConstraint(">= 5.5.0")
	currentVersion, _ := semver.NewVersion(chartVersion)

	if versionWithToolbox.Check(currentVersion) {
		return "toolbox"
	}

	return "task-runner"
}

// NGINXDefaultBackendComponentName returns the component name for NGINXDefaultBackend depending on the Chart version.
// If the Chart version is >= 5.6.0, then it returns "defaultbackend".
// If the Chart version is < 5.6.0, then it returns "default-backend".
// When the list of supported Chart versions are all 5.6.0 or newer, this function
// can be removed and we can use a constant `NGINXDefaultBackendComponentName = "defaultbackend"`.
func NGINXDefaultBackendComponentName(chartVersion string) string {
	versionWithNameChange, _ := semver.NewConstraint(">= 5.6.0")
	currentVersion, _ := semver.NewVersion(chartVersion)

	if versionWithNameChange.Check(currentVersion) {
		return "defaultbackend"
	}

	return "default-backend"
}

func updateCommonLabels(releaseName, componentName string, labels map[string]string) {
	labels["app.kubernetes.io/name"] = releaseName
	labels["app.kubernetes.io/instance"] = fmt.Sprintf("%s-%s", releaseName, componentName)
	labels["app.kubernetes.io/component"] = componentName
	labels["app.kubernetes.io/part-of"] = "gitlab"
	labels["app.kubernetes.io/managed-by"] = "gitlab-operator"
}

func nameWithHashSuffix(name string, adapter CustomResourceAdapter, n int) string {
	suffix := adapter.Hash()[:n]

	if strings.HasSuffix(name, suffix) {
		return name
	}

	return fmt.Sprintf("%s-%s", name, suffix)
}
