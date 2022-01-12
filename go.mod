module gitlab.com/gitlab-org/cloud-native/gitlab-operator

go 1.16

require (
	github.com/Masterminds/semver v1.5.0
	github.com/coreos/prometheus-operator v0.41.1
	github.com/go-logr/logr v0.4.0
	github.com/imdario/mergo v0.3.12
	github.com/jetstack/cert-manager v1.6.1
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.15.0
	github.com/openshift/api v3.9.0+incompatible
	github.com/pkg/errors v0.9.1
	gopkg.in/yaml.v2 v2.4.0
	helm.sh/helm/v3 v3.7.0
	k8s.io/api v0.22.2
	k8s.io/apimachinery v0.22.2
	k8s.io/client-go v0.22.2
	k8s.io/kubectl v0.22.1
	sigs.k8s.io/controller-runtime v0.10.1
	sigs.k8s.io/yaml v1.2.0
)
