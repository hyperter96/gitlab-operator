package settings

import (
	"fmt"
	"net/http"
	"os"

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

	// AppServiceAccount is the name of the ServiceAccount that is used by the application.
	// The default value is "gitlab-manager". Use GITLAB_APP_SERVICE_ACCOUNT environment
	// variable to change it.
	AppServiceAccount = "gitlab-app"

	// NGINXServiceAccount is the name of the ServiceAccount that is used by NGINX.
	// The default value is "gitlab-nginx-ingress". Use NGINX_SERVICE_ACCOUNT environment
	// variable to change it.
	NGINXServiceAccount = "gitlab-nginx-ingress"

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

	KubeVersion *chartutil.KubeVersion = nil
)

const (
	envHelmChartsDirectory   = "HELM_CHARTS"
	envManagerServiceAccount = "GITLAB_MANAGER_SERVICE_ACCOUNT"
	envAppServiceAccount     = "GITLAB_APP_SERVICE_ACCOUNT"
	envNGINXServiceAccount   = "NGINX_SERVICE_ACCOUNT"
	envKubeVersion           = "GITLAB_OPERATOR_KUBERNETES_VERSION"
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

	appServiceAccount := os.Getenv(envAppServiceAccount)
	if appServiceAccount != "" {
		AppServiceAccount = appServiceAccount
	}

	nginxServiceAccount := os.Getenv(envNGINXServiceAccount)
	if nginxServiceAccount != "" {
		NGINXServiceAccount = nginxServiceAccount
	}

	kubeVersionStr := os.Getenv(envKubeVersion)
	if kubeVersionStr != "" {
		KubeVersion, _ = chartutil.ParseKubeVersion(kubeVersionStr)
	}
}
