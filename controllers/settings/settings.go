package settings

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"helm.sh/helm/v3/pkg/chartutil"
)

var (
	// HelmChartsDirectory is the directory that contains all the bundled Charts.
	// The default value is "/charts". Use HELM_CHARTS environment variable to change it.
	HelmChartsDirectory = "/charts"

	// ManagerServiceAccount is the name of the ServiceAccount that is used by the manager.
	// The default value is "gitlab-manager". Use GITLAB_MANAGER_SERVICE_ACCOUNT environment
	// variable to change it.
	ManagerServiceAccount = "gitlab-manager"

	// AppAnyUIDServiceAccount is the name of the ServiceAccount that is used by GitLab components
	// that can run under the 'anyuid' SecurityContextConstraint.
	// The default value is "gitlab-app-anyuid". Use GITLAB_APP_ANYUID_SERVICE_ACCOUNT environment
	// variable to change it.
	AppAnyUIDServiceAccount = "gitlab-app-anyuid"

	// AppNonRootServiceAccount is the name of the ServiceAccount that is used by GitLab components
	// that can run under the 'nonroot' SecurityContextConstraint.
	// The default value is "gitlab-app-nonroot". Use GITLAB_APP_NONROOT_SERVICE_ACCOUNT environment
	// variable to change it.
	AppNonRootServiceAccount = "gitlab-app-nonroot"

	// NGINXServiceAccount is the name of the ServiceAccount that is used by NGINX.
	// The default value is "gitlab-nginx-ingress". Use NGINX_SERVICE_ACCOUNT environment
	// variable to change it.
	NGINXServiceAccount = "gitlab-nginx-ingress"

	// PrometheusServiceAccount is the name of the ServiceAccount that is used by Prometheus.
	// The default value is "gitlab-prometheus-server". Use PROMETHEUS_SERVICE_ACCOUNT environment
	// variable to change it.
	PrometheusServiceAccount = "gitlab-prometheus-server"

	// HealthProbeBindAddress returns the address for hosting health probes.
	HealthProbeBindAddress = ":6060"

	// LivenessEndpointName returns the endpoint name for the liveness probe.
	LivenessEndpointName = "/liveness"

	// ReadinessEndpointName returns the endpoint name for the readiness probe.
	ReadinessEndpointName = "/readiness"

	// AliveStatus returns an error if not alive, and nil if alive.
	AliveStatus = fmt.Errorf("not alive")

	// ReadyStatus returns an error if not ready, and nil if ready.
	ReadyStatus = fmt.Errorf("not ready")

	// HealthzCheck returns the checker.
	HealthzCheck = func(_ *http.Request) error { return AliveStatus }
	ReadyzCheck  = func(_ *http.Request) error { return ReadyStatus }

	DefaultKubeVersion     *chartutil.KubeVersion = nil
	DefaultKubeAPIVersions chartutil.VersionSet   = chartutil.VersionSet{}
)

const (
	envHelmChartsDirectory      = "HELM_CHARTS"
	envManagerServiceAccount    = "GITLAB_MANAGER_SERVICE_ACCOUNT"
	envAppAnyUIDServiceAccount  = "GITLAB_APP_ANYUID_SERVICE_ACCOUNT"
	envAppNonRootServiceAccount = "GITLAB_APP_NONROOT_SERVICE_ACCOUNT"
	envNGINXServiceAccount      = "NGINX_SERVICE_ACCOUNT"
	envPrometheusServiceAccount = "PROMETHEUS_SERVICE_ACCOUNT"
	envKubeVersion              = "GITLAB_OPERATOR_KUBERNETES_VERSION"
	envKubeAPIVersions          = "GITLAB_OPERATOR_KUBERNETES_API_VERSIONS"
)

// Load reads Operator settings from environment variables.
func Load() {
	helmChartsDirectory := os.Getenv(envHelmChartsDirectory)
	if helmChartsDirectory != "" {
		HelmChartsDirectory = helmChartsDirectory
	}

	mgrServiceAccount := os.Getenv(envManagerServiceAccount)
	if mgrServiceAccount != "" {
		ManagerServiceAccount = mgrServiceAccount
	}

	appAnyUIDServiceAccount := os.Getenv(envAppAnyUIDServiceAccount)
	if appAnyUIDServiceAccount != "" {
		AppAnyUIDServiceAccount = appAnyUIDServiceAccount
	}

	appNonRootServiceAccount := os.Getenv(envAppNonRootServiceAccount)
	if appNonRootServiceAccount != "" {
		AppNonRootServiceAccount = appNonRootServiceAccount
	}

	nginxServiceAccount := os.Getenv(envNGINXServiceAccount)
	if nginxServiceAccount != "" {
		NGINXServiceAccount = nginxServiceAccount
	}

	prometheusServiceAccount := os.Getenv(envPrometheusServiceAccount)
	if prometheusServiceAccount != "" {
		PrometheusServiceAccount = prometheusServiceAccount
	}

	kubeVersionStr := os.Getenv(envKubeVersion)
	if kubeVersionStr != "" {
		DefaultKubeVersion, _ = chartutil.ParseKubeVersion(kubeVersionStr)
	}

	kubeAPIVersionsStr := os.Getenv(envKubeAPIVersions)
	if kubeAPIVersionsStr != "" {
		DefaultKubeAPIVersions = strings.Split(kubeAPIVersionsStr, ",")
	}
}
