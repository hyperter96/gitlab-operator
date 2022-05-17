package gitlab

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/support"
)

var _ = Describe("Getting default values from charts", func() {
	// None of these have defaults configured in the CustomResourceAdapter,
	// so their return values should come from the chart defaults.
	tests := map[string]interface{}{
		"gitlab.gitlab-exporter.enabled":                  true,
		"gitlab.gitlab-shell.enabled":                     true,
		"gitlab.mailroom.enabled":                         true,
		"gitlab.migrations.enabled":                       true,
		"gitlab.sidekiq.enabled":                          true,
		"gitlab.toolbox.backups.cron.enabled":             false,
		"gitlab.toolbox.backups.cron.persistence.enabled": false,
		"gitlab.toolbox.enabled":                          true,
		"gitlab.webservice.enabled":                       true,
		"global.appConfig.incomingEmail.enabled":          false,
		"global.appConfig.incomingEmail.password.secret":  "",
		"global.gitaly.enabled":                           true,
		"global.hosts.domain":                             "example.com",
		"global.ingress.configureCertmanager":             true,
		"global.ingress.provider":                         "nginx",
		"global.kas.enabled":                              false,
		"global.pages.enabled":                            false,
		"nginx-ingress.enabled":                           true,
		"postgresql.isntall":                              true,
		"redis.install":                                   true,
		"registry.enabled":                                true,
	}

	for key, expectedValue := range tests {
		// Must assign a copy of the loop variable to a local variable:
		// https://onsi.github.io/ginkgo/#dynamically-generating-specs
		key := key
		expectedValue := expectedValue

		chartValues := support.Values{}
		_ = chartValues.SetValue(key, expectedValue)

		mockGitLab := CreateMockGitLab(releaseName, namespace, chartValues)
		adapter := CreateMockAdapter(mockGitLab)
		template, templateErr := GetTemplate(adapter)
		getValue, getValueErr := adapter.Values().GetValue(key)

		When(fmt.Sprintf("getting value for %s", key), func() {
			It("Should render the template", func() {
				Expect(templateErr).To(BeNil())
				Expect(template).NotTo(BeNil())
			})

			It(fmt.Sprintf("should get %s for %s", expectedValue, key), func() {
				Expect(getValueErr).To(BeNil())
				Expect(getValue).To(Equal(expectedValue))
			})
		})
	}
})
