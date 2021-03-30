package gitlab

import (
	"os"
	"strconv"
	"strings"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	// SharedSecretsJobDefaultTimeout is the default timeout to wait for Shared Secrets job to finish.
	SharedSecretsJobDefaultTimeout = 300 * time.Second
)

// SharedSecretsConfigMap returns the ConfigMaps of Shared Secret component.
func SharedSecretsConfigMap(adapter CustomResourceAdapter) (*corev1.ConfigMap, error) {
	template, err := GetTemplate(adapter)

	if err != nil {
		return nil, err
	}

	cfgMap := template.Query().ConfigMapByComponent(SharedSecretsComponentName)

	return cfgMap, nil
}

// SharedSecretsJob returns the Job for Shared Secret component.
func SharedSecretsJob(adapter CustomResourceAdapter) (*batchv1.Job, error) {
	template, err := GetTemplate(adapter)

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
func SelfSignedCertsJob(adapter CustomResourceAdapter) (*batchv1.Job, error) {
	template, err := GetTemplate(adapter)

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

// SharedSecretsResources returns Kubernetes resources for running shared secrets job.
func SharedSecretsResources(adapter CustomResourceAdapter) (*corev1.ConfigMap, *batchv1.Job, error) {

	cfgMap, err := SharedSecretsConfigMap(adapter)
	if err != nil {
		return nil, nil, err
	}
	if cfgMap == nil {
		return nil, nil,
			errors.NewNotFound(schema.ParseGroupResource("configmaps"), SharedSecretsComponentName)
	}

	job, err := SharedSecretsJob(adapter)
	if err != nil {
		return nil, nil, err
	}
	if job == nil {
		return nil, nil,
			errors.NewNotFound(schema.ParseGroupResource("jobs.batch"), SharedSecretsComponentName)
	}

	return cfgMap, job, nil
}

// SharedSecretsJobTimeout returns the timeout for shared secrets job to finish.
func SharedSecretsJobTimeout() time.Duration {
	s := os.Getenv("GITLAB_OPERATOR_SHARED_SECRETS_JOB_TIMEOUT")
	if s != "" {
		i, err := strconv.ParseInt(s, 10, 64)
		if err == nil {
			return time.Duration(i) * time.Second
		}
	}
	return SharedSecretsJobDefaultTimeout
}

// SharedSecretsJobWaitPeriod returns the wait time next check of shared secrets job status.
func SharedSecretsJobWaitPeriod(timeout, elapsed time.Duration) time.Duration {
	return time.Duration(timeout / 100)
}
