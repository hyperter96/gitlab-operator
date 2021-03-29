package gitlab

import (
	"strings"

	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/helpers"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

// SharedSecretsConfigMap returns the ConfigMaps of Shared Secret component.
func SharedSecretsConfigMap(adapter helpers.CustomResourceAdapter) (*corev1.ConfigMap, error) {
	template, err := helpers.GetTemplate(adapter)

	if err != nil {
		return nil, err
	}

	cfgMap := template.Query().ConfigMapByComponent(SharedSecretsComponentName)

	return cfgMap, nil
}

// SharedSecretsJob returns the Job for Shared Secret component.
func SharedSecretsJob(adapter helpers.CustomResourceAdapter) (*batchv1.Job, error) {
	template, err := helpers.GetTemplate(adapter)

	if err != nil {
		return nil, err
	}

	jobs := template.Query().JobsByLabels(map[string]string{
		"app": SharedSecretsComponentName,
	})

	for _, j := range jobs {
		if !strings.HasSuffix(j.ObjectMeta.Name, "-selfsign") {
			return j, nil
		}
	}
	return nil, nil
}

// SelfSignedCertsJob returns the Job for Self Signed Certificates component.
func SelfSignedCertsJob(adapter helpers.CustomResourceAdapter) (*batchv1.Job, error) {
	template, err := helpers.GetTemplate(adapter)

	if err != nil {
		return nil, err
	}

	jobs := template.Query().JobsByLabels(map[string]string{
		"app": SharedSecretsComponentName,
	})

	for _, j := range jobs {
		if strings.HasSuffix(j.ObjectMeta.Name, "-selfsign") {
			return j, nil
		}
	}
	return nil, nil
}
