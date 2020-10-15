module gitlab.com/gitlab-org/gl-openshift/gitlab-operator

go 1.13

require (
	github.com/coreos/prometheus-operator v0.41.1
	github.com/go-logr/logr v0.1.0
	github.com/imdario/mergo v0.3.9
	github.com/jetstack/cert-manager v0.16.1
	github.com/nginxinc/nginx-ingress-operator v0.0.6
	github.com/onsi/ginkgo v1.12.1
	github.com/onsi/gomega v1.10.1
	github.com/openshift/api v3.9.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/prometheus/common v0.10.0
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	helm.sh/helm/v3 v3.3.0
	k8s.io/api v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/kubectl v0.18.6
	sigs.k8s.io/controller-runtime v0.6.2
	sigs.k8s.io/yaml v1.2.0
)

replace k8s.io/client-go => k8s.io/client-go v0.18.6
