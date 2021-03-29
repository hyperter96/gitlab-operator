package gitlab

import (
	"fmt"

	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/helpers"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

// MigrationsConfigMap returns the ConfigMaps of Migrations component.
func MigrationsConfigMap(adapter helpers.CustomResourceAdapter) *corev1.ConfigMap {
	var result *corev1.ConfigMap
	template, err := helpers.GetTemplate(adapter)

	if err != nil {
		return result
		/* WARNING: This should return an error instead. */
	}

	result = template.Query().ConfigMapByName(
		fmt.Sprintf("%s-%s", adapter.ReleaseName(), MigrationsComponentName))

	return result
}

// MigrationsJob returns the Job for Migrations component.
func MigrationsJob(adapter helpers.CustomResourceAdapter) (*batchv1.Job, error) {
	template, err := helpers.GetTemplate(adapter)

	if err != nil {
		return nil, err
	}

	result := template.Query().JobByComponent(MigrationsComponentName)

	return result, nil
}
