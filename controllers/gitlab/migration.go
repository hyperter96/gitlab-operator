package gitlab

import (
	"fmt"
	"os"
	"strconv"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	// MigrationsJobDefaultTimeout is the default timeout to wait for Migrations job to finish.
	MigrationsJobDefaultTimeout = 300 * time.Second
)

// MigrationsConfigMap returns the ConfigMaps of Migrations component.
func MigrationsConfigMap(adapter CustomResourceAdapter) *corev1.ConfigMap {
	var result *corev1.ConfigMap
	template, err := GetTemplate(adapter)

	if err != nil {
		return result
		/* WARNING: This should return an error instead. */
	}

	result = template.Query().ConfigMapByName(
		fmt.Sprintf("%s-%s", adapter.ReleaseName(), MigrationsComponentName))

	return result
}

// MigrationsJob returns the Job for Migrations component.
func MigrationsJob(adapter CustomResourceAdapter) (*batchv1.Job, error) {
	template, err := GetTemplate(adapter)

	if err != nil {
		return nil, err
	}

	result := template.Query().JobByComponent(MigrationsComponentName)

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
