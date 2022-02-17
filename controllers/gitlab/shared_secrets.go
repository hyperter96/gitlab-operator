package gitlab

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
)

const (
	// SharedSecretsJobDefaultTimeout is the default timeout to wait for Shared Secrets job to finish.
	SharedSecretsJobDefaultTimeout = 300 * time.Second
)

// SharedSecretsConfigMap returns the ConfigMaps of Shared Secret component.
func SharedSecretsConfigMap(adapter CustomResourceAdapter, template helm.Template) (client.Object, error) {
	cfgMapName := fmt.Sprintf("%s-%s", adapter.ReleaseName(), SharedSecretsComponentName)
	cfgMap := template.Query().ObjectByKindAndName(ConfigMapKind, cfgMapName)

	return cfgMap, nil
}

// SharedSecretsJob returns the Job for Shared Secret component.
func SharedSecretsJob(adapter CustomResourceAdapter, template helm.Template) (client.Object, error) {
	jobs := template.Query().ObjectsByKindAndLabels(JobKind, map[string]string{
		"app": GitLabComponentName,
	})

	namePrefix := fmt.Sprintf("%s-%s", adapter.ReleaseName(), SharedSecretsComponentName)
	for _, j := range jobs {
		if strings.HasPrefix(j.GetName(), namePrefix) && !strings.HasSuffix(j.GetName(), "-selfsign") {
			return j, nil
		}
	}

	return nil, nil
}

// SelfSignedCertsJob returns the Job for Self Signed Certificates component.
func SelfSignedCertsJob(adapter CustomResourceAdapter, template helm.Template) (client.Object, error) {
	jobs := template.Query().ObjectsByKindAndLabels(JobKind, map[string]string{
		"app": GitLabComponentName,
	})

	namePrefix := fmt.Sprintf("%s-%s", adapter.ReleaseName(), SharedSecretsComponentName)
	for _, j := range jobs {
		if strings.HasPrefix(j.GetName(), namePrefix) && strings.HasSuffix(j.GetName(), "-selfsign") {
			return j, nil
		}
	}

	return nil, nil
}

// SharedSecretsResources returns Kubernetes resources for running shared secrets job.
func SharedSecretsResources(adapter CustomResourceAdapter, template helm.Template) (client.Object, client.Object, error) {
	cfgMap, err := SharedSecretsConfigMap(adapter, template)
	if err != nil {
		return cfgMap, nil, err
	}

	job, err := SharedSecretsJob(adapter, template)
	if err != nil {
		return cfgMap, job, err
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
