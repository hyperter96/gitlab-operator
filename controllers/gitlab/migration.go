package gitlab

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
)

const (
	// MigrationsJobDefaultTimeout is the default timeout to wait for Migrations job to finish.
	MigrationsJobDefaultTimeout = 300 * time.Second

	gitlabMigrationsEnabled  = "gitlab.migrations.enabled"
	migrationsEnabledDefault = true
)

// MigrationsEnabled returns `true` if enabled and `false` if not.
func MigrationsEnabled(adapter CustomResourceAdapter) bool {
	return adapter.Values().GetBool(gitlabMigrationsEnabled, migrationsEnabledDefault)
}

// MigrationsConfigMap returns the ConfigMaps of Migrations component.
func MigrationsConfigMap(adapter CustomResourceAdapter, template helm.Template) client.Object {
	return template.Query().ObjectByKindAndName(ConfigMapKind,
		fmt.Sprintf("%s-%s", adapter.ReleaseName(), MigrationsComponentName))
}

// MigrationsJob returns the Job for Migrations component.
func MigrationsJob(adapter CustomResourceAdapter, template helm.Template) (client.Object, error) {
	result := template.Query().ObjectByKindAndComponent(JobKind, MigrationsComponentName)
	result.SetName(nameWithHashSuffix(result.GetName(), adapter, 3))

	return result, nil
}

// MigrationsJobTimeout returns the timeout for shared secrets job to finish.
func MigrationsJobTimeout() time.Duration {
	s := os.Getenv("GITLAB_OPERATOR_MIGRATIONS_JOB_TIMEOUT")
	if s != "" {
		i, err := strconv.ParseInt(s, 10, 64)
		if err == nil {
			return time.Duration(i) * time.Second
		}
	}

	return MigrationsJobDefaultTimeout
}
