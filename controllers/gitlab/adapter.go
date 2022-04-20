package gitlab

import (
	"fmt"
	"hash/fnv"
	"strconv"
	"strings"

	gitlabv1beta1 "gitlab.com/gitlab-org/cloud-native/gitlab-operator/api/v1beta1"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/settings"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/resource"

	"github.com/Masterminds/semver/v3"
	"github.com/mitchellh/copystructure"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
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

	// StatusVersion returns the version of the GitLab chart that the GitLab
	// Custom Resource is actively running.
	StatusVersion() string

	// IsUpgrade returns `true` if StatusVersion is set and is not equal to
	// ChartVersion. Otherwise, it returns `false`.
	IsUpgrade() bool

	// Values returns the set of values that will be used the render GitLab chart.
	Values() resource.Values

	// ResetValues re-initializes the values of the adapter with the values of
	// GitLab resource and Operator defaults.
	ResetValues(resource *gitlabv1beta1.GitLab)

	// UpdateValues coalesces all the values in the Chart with the current values.
	// This is to ensure that Chart default values are populated as well.
	UpdateValues(chart *chart.Chart) error
}

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
  toolbox:
    backups:
      objectStorage:
        config:
          secret: $ToolboxConnectionSecretName
          key: config
    common:
      labels:
        app.kubernetes.io/component: toolbox
        app.kubernetes.io/instance: $ReleaseName-toolbox
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
  mailroom:
    common:
    labels:
      app.kubernetes.io/component: mailroom
      app.kubernetes.io/instance: $ReleaseName-mailroom
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

shared-secrets:
  serviceAccount:
    create: false
    name: $ManagerServiceAccount
  securityContext:
    runAsUser: ''
    fsGroup: ''

global:
  common:
    labels:
      app.kubernetes.io/name: $ReleaseName
      app.kubernetes.io/part-of: gitlab
      app.kubernetes.io/managed-by: gitlab-operator
  image:
    pullPolicy: IfNotPresent
  ingress:
    apiVersion: networking.k8s.io/v1
    annotations:
      $GlobalIngressAnnotations
  serviceAccount:
    enabled: true
    create: false
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

nginx-ingress:
  rbac:
    create: false
  serviceAccount:
    name: $NGINXServiceAccount
  controller:
    service:
      loadBalancerIP: $GlobalHostsExternalIP
  defaultBackend:
    serviceAccount:
      name: $AppServiceAccount
`

var defaultValuesMinio string = `
global:
  minio:
    enabled: false
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
  registry:
    bucket: registry

registry:
  storage:
    secret: $RegistryConnectionSecretName
    key: config
    redirect:
      disable: $RegistryMinioRedirect
`

// NewCustomResourceAdapter returns a new adapter for the provided GitLab instance.
func NewCustomResourceAdapter(gitlab *gitlabv1beta1.GitLab) CustomResourceAdapter {
	result := &populatingAdapter{
		resource: gitlab,
	}
	result.ResetValues(gitlab)

	return result
}

type populatingAdapter struct {
	resource  *gitlabv1beta1.GitLab
	values    resource.Values
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

func (a *populatingAdapter) StatusVersion() string {
	return a.resource.Status.Version
}

func (a *populatingAdapter) IsUpgrade() bool {
	return a.StatusVersion() != "" && a.StatusVersion() != a.ChartVersion()
}

func (a *populatingAdapter) ReleaseName() string {
	return a.resource.Name
}

func (a *populatingAdapter) Values() resource.Values {
	return a.values
}

func (a *populatingAdapter) ResetValues(resource *gitlabv1beta1.GitLab) {
	if vCopy, err := copystructure.Copy(resource.Spec.Chart.Values.Object); err == nil {
		a.values = vCopy.(map[string]interface{})
	} else {
		a.values = resource.Spec.Chart.Values.Object
	}

	a.populateValues()
	a.hashValues()
}

func (a *populatingAdapter) UpdateValues(chart *chart.Chart) error {
	coalesceValues, err := chartutil.CoalesceValues(chart, a.values)

	if err == nil {
		a.values = coalesceValues.AsMap()
	}

	return err
}

func (a *populatingAdapter) populateValues() {
	a.reference = fmt.Sprintf("%s.%s", a.resource.Name, a.resource.Namespace)

	// Need to pass a default value here as we don't yet have the coalesced values from GetTemplate().
	configureCertmanager := a.values.GetBool("global.ingress.configureCertmanager", true)

	globalIngressAnnotations := "{}"

	if configureCertmanager {
		issuerAnnotation := fmt.Sprintf("cert-manager.io/issuer: %s-issuer", a.ReleaseName())
		acmeAnnotation := "acme.cert-manager.io/http01-edit-in-place: \"true\""
		globalIngressAnnotations = fmt.Sprintf("%s\n      %s", issuerAnnotation, acmeAnnotation)
	}

	globalHostsExternalIP := a.values.GetString("global.hosts.externalIP")

	valuesToUse := strings.NewReplacer(
		"$ReleaseName", a.ReleaseName(),
		"$LocalUser", settings.LocalUser,
		"$AppServiceAccount", settings.AppServiceAccount,
		"$ManagerServiceAccount", settings.ManagerServiceAccount,
		"$ToolboxConnectionSecretName", settings.ToolboxConnectionSecretName,
		"$GlobalIngressAnnotations", globalIngressAnnotations,
		"$NGINXServiceAccount", settings.NGINXServiceAccount,
		"$GlobalHostsExternalIP", globalHostsExternalIP,
	).Replace(defaultValues)

	_ = a.values.AddFromYAML(valuesToUse)

	// Need to pass a default value here as we don't yet have the coalesced values from GetTemplate().
	minioEnabled := a.values.GetBool(globalMinioEnabled, true)
	if minioEnabled {
		minioRedirect := a.values.GetBool("registry.minio.redirect")
		valuesToUse := strings.NewReplacer(
			"$AppConfigConnectionSecretName", settings.AppConfigConnectionSecretName,
			"$RegistryConnectionSecretName", settings.RegistryConnectionSecretName,
			"$RegistryMinioRedirect", strconv.FormatBool(!minioRedirect),
		).Replace(defaultValuesMinio)

		_ = a.values.AddFromYAML(valuesToUse)
	}

	// This is a workaround to account for the fact that our "internal" MinIO is actually
	// implemented as external object storage, meaning `global.minio.enabled` must be
	// set to `false`. If `internalMinioEnabled=true`, then our "internal" MinIO objects
	// will be reconciled, and vice versa.
	_ = a.values.SetValue(internalMinioEnabled, minioEnabled)

	email := a.values.GetString("certmanager-issuer.email")
	if email == "" {
		_ = a.values.SetValue("certmanager-issuer.email", "admin@example.com")
	}

	// Per https://gitlab.com/gitlab-org/cloud-native/gitlab-operator/-/issues/625,
	// the kubectl image tagged 1.18.20 needed to be patched to include an environment
	// variable setting HOME=/tmp/kube to work around throttling issues.
	// This was fixed with https://gitlab.com/gitlab-org/charts/gitlab/-/merge_requests/2486,
	// but that change won't be available until 5.9.4 (or 5.10.x, whichever comes first).
	// Versions prior to 5.9.x use an older version of the kubectl image that do not
	// have this throttling problem.

	needsPatch, err := semver.NewConstraint("5.9.0 - 5.9.3")
	if err != nil {
		return
	}

	currentVersion, err := semver.NewVersion(a.ChartVersion())
	if err != nil {
		return
	}

	if needsPatch.Check(currentVersion) {
		_ = a.values.SetValue("global.kubectl.image.tag", "1.18.20@sha256:f0bc9adaccd131d993fdabf6aa39fe4ff0e22035c2deca20341074c9e2e40a5b")
	}
}

func (a *populatingAdapter) hashValues() {
	hasher := fnv.New64()
	valuesToHash := []([]byte){
		[]byte(a.Namespace()),
		[]byte(a.ReleaseName()),
		[]byte(a.ChartVersion()),
		[]byte(fmt.Sprintf("%s", a.values)),
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
