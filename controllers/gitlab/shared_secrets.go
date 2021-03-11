package gitlab

import (
	"os"
	"strconv"
	"time"

	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/helpers"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	// SharedSecretsJobDefaultTimeout is the default timeout to wait for Shared Secrets job to finish.
	SharedSecretsJobDefaultTimeout = 300 * time.Second
)

// SharedSecretsResources returns Kubernetes resources for running shared secrets job.
func SharedSecretsResources(adapter helpers.CustomResourceAdapter) (*corev1.ConfigMap, *batchv1.Job, error) {

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
