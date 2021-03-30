package gitlab

import (
	"fmt"
	"hash/fnv"
	"strings"

	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/settings"
	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/helm"
)

// CustomResourceAdapter is a wrapper for GitLab Custom Resource. It provides a convenient interface
// to interact with the GitLab instances and guards the controller from its structural changes.
//
// This adapter is immutable and will not update itself after initialization. Therefore, it must be
// created when GitLab Custom Resource changes, e.g. in reconcile loop.
type CustomResourceAdapter interface {
	// Resource returns the reference to the underlaying Custom Resource.
	Resource() *gitlabv1beta1.GitLab

	// Hash generates a hash based on the key parts of a GitLab Custom Resource. The hash can be used
	// to identify changes to the underlaying resource. For example this is useful when rendering a
	// Helm template.
	Hash() string

	// Reference returns a fully qualified name of the associated GitLab Custom Resource. As opposed
	// to Hash this value does not change.
	Reference() string

	// Namespace returns the namespace in which the GitLab instance must be deployed. When Operator
	// is scoped to
	// a namespace this must be equal to the namespace of the Operator.
	Namespace() string

	// ReleaseName returns the name of the GitLab instance that must be deployed. This will be used
	// as a qualifier to distinguish between multiple GitLab instances in a namespace.
	ReleaseName() string

	// ChartVersion returns the version of GitLab chart that must be used to deploy this GitLab
	// instance.
	ChartVersion() string

	// Values returns the set of values that will be used the render GitLab chart.
	Values() helm.Values
}

//TODO: Use go:embed instead
var defaultValues string = `
certmanager:
  install: false
gitlab-runner:
  install: false
gitlab:
  gitaly:
    common:
      labels:
        app.kubernetes.io/component: gitaly
        app.kubernetes.io/instance: $ReleaseName-gitaly
    securityContext:
      runAsUser: $LocalUser
      fsGroup: $LocalUser
  gitlab-exporter:
    common:
      labels:
        app.kubernetes.io/component: gitlab-exporter
        app.kubernetes.io/instance: $ReleaseName-gitlab-exporter
    securityContext:
      runAsUser: $LocalUser
      fsGroup: $LocalUser
  gitlab-shell:
    common:
      labels:
        app.kubernetes.io/component: gitlab-shell
        app.kubernetes.io/instance: $ReleaseName-gitlab-shell
    securityContext:
      runAsUser: $LocalUser
      fsGroup: $LocalUser
    service:
      type: ''
  migrations:
    common:
      labels:
        app.kubernetes.io/component: migrations
        app.kubernetes.io/instance: $ReleaseName-migrations
  sidekiq:
    common:
      labels:
        app.kubernetes.io/component: sidekiq
        app.kubernetes.io/instance: $ReleaseName-sidekiq
    securityContext:
      runAsUser: $LocalUser
      fsGroup: $LocalUser
  task-runner:
    backups:
      objectStorage:
        config:
          secret: $TaskRunnerConnectionSecretName
          key: config
    common:
      labels:
        app.kubernetes.io/component: task-runner
        app.kubernetes.io/instance: $ReleaseName-task-runner
    securityContext:
      runAsUser: $LocalUser
      fsGroup: $LocalUser
  webservice:
    common:
      labels:
        app.kubernetes.io/component: webservice
        app.kubernetes.io/instance: $ReleaseName-webservice
    securityContext:
      runAsUser: $LocalUser
      fsGroup: $LocalUser

registry:
  common:
    labels:
      app.kubernetes.io/component: registry
      app.kubernetes.io/instance: $ReleaseName-registry
  securityContext:
    runAsUser: $LocalUser
    fsGroup: $LocalUser
  storage:
    secret: $RegistryConnectionSecretName
    key: config

shared-secrets:
  serviceAccount:
    create: false
    name: $ManagerServiceAccount
  securityContext:
    runAsUser: ''
    fsGroup: ''

global:
  appConfig:
    object_store:
      enabled: true
      connection:
        secret: $AppConfigConnectionSecretName
        key: connection
    artifacts:
      bucket: gitlab-artifacts
    backups:
      bucket: gitlab-backups
      tmpBucket: tmp
    externalDiffs:
      bucket: gitlab-mr-diffs
    lfs:
      bucket: git-lfs
    packages:
      bucket: gitlab-packages
    pseudonymizer:
      bucket: gitlab-pseudo
    uploads:
      bucket: gitlab-uploads
  common:
    labels:
      app.kubernetes.io/name: $ReleaseName
      app.kubernetes.io/part-of: gitlab
      app.kubernetes.io/managed-by: gitlab-operator
  imagePullPolicy: IfNotPresent
  ingress:
    configureCertmanager: false
  minio:
    enabled: false
  registry:
    bucket: registry
  serviceAccount:
    enabled: true,
    create: false,
    name: $AppServiceAccount

redis:
  master:
    statefulset:
      labels:
        app.kubernetes.io/name: $ReleaseName
        app.kubernetes.io/part-of: gitlab
        app.kubernetes.io/managed-by: gitlab-operator
        app.kubernetes.io/component: redis
        app.kubernetes.io/instance: $ReleaseName-redis
  serviceAccount:
    name: $AppServiceAccount
  securityContext:
    runAsUser: $LocalUser
    fsGroup: $LocalUser

postgresql:
  serviceAccount:
    enabled: true
    name: $AppServiceAccount
  securityContext:
    runAsUser: $LocalUser
    fsGroup: $LocalUser  
`

// NewCustomResourceAdapter returns a new adapter for the provided GitLab instance.
func NewCustomResourceAdapter(gitlab *gitlabv1beta1.GitLab) CustomResourceAdapter {
	result := &populatingAdapter{
		resource: gitlab,
		values:   helm.FromMap(gitlab.Spec.Chart.Values.Object),
	}
	result.populateValues()
	result.hashValues()
	return result
}

type populatingAdapter struct {
	resource  *gitlabv1beta1.GitLab
	values    helm.Values
	hash      string
	reference string
}

func (a *populatingAdapter) Resource() *gitlabv1beta1.GitLab {
	return a.resource
}

func (a *populatingAdapter) Hash() string {
	return a.hash
}

func (a *populatingAdapter) Reference() string {
	return a.reference
}

func (a *populatingAdapter) Namespace() string {
	return a.resource.Namespace
}

func (a *populatingAdapter) ChartVersion() string {
	return a.resource.Spec.Chart.Version
}

func (a *populatingAdapter) ReleaseName() string {
	return a.resource.Name
}

func (a *populatingAdapter) Values() helm.Values {
	return a.values
}

func (a *populatingAdapter) populateValues() {
	a.reference = fmt.Sprintf("%s.%s", a.resource.Name, a.resource.Namespace)

	valuesToUse := strings.NewReplacer(
		"$ReleaseName", a.ReleaseName(),
		"$LocalUser", settings.LocalUser,
		"$AppServiceAccount", settings.AppServiceAccount,
		"$ManagerServiceAccount", settings.ManagerServiceAccount,
		"$AppConfigConnectionSecretName", settings.AppConfigConnectionSecretName,
		"$RegistryConnectionSecretName", settings.RegistryConnectionSecretName,
		"$TaskRunnerConnectionSecretName", settings.TaskRunnerConnectionSecretName,
	).Replace(defaultValues)

	a.values.AddFromYAML([]byte(valuesToUse))

	email, err := GetStringValue(a.Values(), "certmanager-issuer.email")
	if err != nil || email == "" {
		a.values.SetValue("certmanager-issuer.email", "admin@example.com")
	}
}

func (a *populatingAdapter) hashValues() {
	hasher := fnv.New64()
	valuesToHash := []([]byte){
		[]byte(a.Namespace()),
		[]byte(a.ReleaseName()),
		[]byte(a.ChartVersion()),
		[]byte(fmt.Sprintf("%s", a.Values())),
	}
	valuesHashed := 0

	for _, v := range valuesToHash {
		_, err := hasher.Write(v)

		if err == nil {
			valuesHashed++
		}
	}

	if valuesHashed == 0 {
		a.hash = fmt.Sprintf("%s/%s", a.Reference(), a.ChartVersion())
	}

	a.hash = fmt.Sprintf("%x", hasher.Sum64())
}