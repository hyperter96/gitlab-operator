package settings

import (
	"os"
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
)

const (
	envHelmChartsDirectory   = "HELM_CHARTS"
	envManagerServiceAccount = "GITLAB_MANAGER_SERVICE_ACCOUNT"
	envAppServiceAccount     = "GITLAB_APP_SERVICE_ACCOUNT"
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
}
