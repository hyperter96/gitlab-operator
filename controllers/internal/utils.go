package internal

import (
	"strings"

	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/gitlab"
)

func getGitlabURL(adapter gitlab.CustomResourceAdapter) string {
	url, err := gitlab.GetStringValue(adapter.Values(), "global.hosts.gitlab.name")

	if err == nil && url != "" {
		return url
	}

	return "gitlab.example.com"
}

func getRegistryURL(adapter gitlab.CustomResourceAdapter) string {
	url, err := gitlab.GetStringValue(adapter.Values(), "global.hosts.registry.name")

	if err == nil && url != "" {
		return url
	}

	return "registry.example.com"
}

func domainNameOnly(url string) string {
	if strings.Contains(url, "://") {
		return strings.Split(url, "://")[1]
	}

	return url
}

func getName(cr, component string) string {
	return strings.Join([]string{cr, component}, "-")
}
