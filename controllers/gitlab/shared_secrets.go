package gitlab

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// SharedSecretsJobDefaultTimeout is the default timeout to wait for Shared Secrets job to finish.
	SharedSecretsJobDefaultTimeout = 300 * time.Second

	sharedSecretsEnabled        = "shared-secrets.enabled"
	sharedSecretsEnabledDefault = true
)

// SharedSecretsEnabled returns `true` if the Shared Secrets component is enabled, and `false` if not.
func SharedSecretsEnabled(adapter CustomResourceAdapter) bool {
	enabled := adapter.Values().GetBool(sharedSecretsEnabled, sharedSecretsEnabledDefault)

	return enabled
}

// SelfSignedCertsEnabled returns `true` if the self-signed certificates component is enabled, and `false` if not.
func SelfSignedCertsEnabled(adapter CustomResourceAdapter) bool {
	sharedSecretsEnabled := SharedSecretsEnabled(adapter)
	configureCertmanager := adapter.Values().GetBool("global.ingress.configureCertmanager", true)
	globalTLSConfigured, _ := adapter.Values().GetValue("global.ingress.tls")

	if sharedSecretsEnabled && !configureCertmanager && globalTLSConfigured == nil {
		return true
	}

	return false
}

// SharedSecretsConfigMap returns the ConfigMaps of Shared Secret component.
func SharedSecretsConfigMap(adapter CustomResourceAdapter) (client.Object, error) {
	template, err := GetTemplate(adapter)

	if err != nil {
		return nil, err
	}

	cfgMapName := fmt.Sprintf("%s-%s", adapter.ReleaseName(), SharedSecretsComponentName)
	cfgMap := template.Query().ObjectByKindAndName(ConfigMapKind, cfgMapName)

	return cfgMap, nil
}

// SharedSecretsJob returns the Job for Shared Secret component.
func SharedSecretsJob(adapter CustomResourceAdapter) (client.Object, error) {
	template, err := GetTemplate(adapter)

	if err != nil {
		return nil, err
	}

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
func SelfSignedCertsJob(adapter CustomResourceAdapter) (client.Object, error) {
	template, err := GetTemplate(adapter)

	if err != nil {
		return nil, err
	}

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
func SharedSecretsResources(adapter CustomResourceAdapter) (client.Object, client.Object, error) {
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
