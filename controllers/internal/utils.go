package internal

import (
	"fmt"
	"strings"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/gitlab"
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

func getMinioURL(adapter gitlab.CustomResourceAdapter) string {
	name, err := gitlab.GetStringValue(adapter.Values(), "global.hosts.minio.name")
	if err != nil {
		// Parameter is optional. Safe to continue.
	}

	if name != "" {
		return name
	}

	hostSuffix, err := gitlab.GetStringValue(adapter.Values(), "global.hosts.hostSuffix")
	if err != nil {
		// Parameter is optional. Safe to continue.
	}

	domain, err := gitlab.GetStringValue(adapter.Values(), "global.hosts.domain")
	if err != nil {
		domain = "example.com"
	}

	if hostSuffix != "" {
		return fmt.Sprintf("minio-%s.%s", hostSuffix, domain)
	}

	return fmt.Sprintf("minio.%s", domain)
}

func getName(cr, component string) string {
	return strings.Join([]string{cr, component}, "-")
}
