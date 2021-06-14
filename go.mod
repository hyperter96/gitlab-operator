module gitlab.com/gitlab-org/cloud-native/gitlab-operator

go 1.14

require (
	github.com/Masterminds/semver v1.5.0
	github.com/coreos/prometheus-operator v0.41.1
	github.com/go-logr/logr v0.3.0
	github.com/imdario/mergo v0.3.11
	github.com/jetstack/cert-manager v0.15.2
	github.com/nxadm/tail v1.4.5 // indirect
	github.com/onsi/ginkgo v1.14.2
	github.com/onsi/gomega v1.10.2
	github.com/openshift/api v3.9.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/prometheus/common v0.10.0
	gopkg.in/yaml.v2 v2.3.0
	helm.sh/helm/v3 v3.5.3
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	k8s.io/kubectl v0.20.2
	sigs.k8s.io/controller-runtime v0.8.3
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible
	k8s.io/apimachinery => k8s.io/apimachinery v0.20.2
	k8s.io/client-go => k8s.io/client-go v0.20.2
)
