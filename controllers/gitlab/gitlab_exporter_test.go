package gitlab

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/gitlab/component"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/support"
)

const (
	gitlabExporterEnabled = "gitlab.gitlab-exporter.enabled"
)

var _ = Describe("Enabling or disabling GitLab Exporter", func() {
	chartValuesDefault := support.Values{}

	chartValuesEnabled := support.Values{}
	_ = chartValuesEnabled.SetValue(gitlabExporterEnabled, true)

	chartValuesDisabled := support.Values{}
	_ = chartValuesDisabled.SetValue(gitlabExporterEnabled, false)

	tests := map[string]struct {
		chartValues support.Values
		expected    bool
	}{
		"enabled (default)": {chartValues: chartValuesDefault, expected: true},
		"enabled":           {chartValues: chartValuesEnabled, expected: true},
		"disabled":          {chartValues: chartValuesDisabled, expected: false},
	}

	for name, test := range tests {
		// Must assign a copy of the loop variable to a local variable:
		// https://onsi.github.io/ginkgo/#dynamically-generating-specs
		name := name
		test := test

		When(name, func() {
			mockGitLab := CreateMockGitLab(releaseName, namespace, test.chartValues)
			adapter := CreateMockAdapter(mockGitLab)
			template, err := GetTemplate(adapter)
			enabled := adapter.WantsComponent(component.GitLabExporter)

			It("Should render the template", func() {
				Expect(err).To(BeNil())
				Expect(template).NotTo(BeNil())
			})

			It(fmt.Sprintf("Should have %s GitLab Exporter", name), func() {
				Expect(enabled).To(Equal(test.expected))
			})
		})
	}
})
