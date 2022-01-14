package gitlab

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/resource"
)

var _ = Describe("Enabling or disabling internal MinIO", func() {
	chartValuesDefault := resource.Values{}

	chartValuesEnabled := resource.Values{}
	_ = chartValuesEnabled.SetValue(globalMinioEnabled, true)

	chartValuesDisabled := resource.Values{}
	chartValuesDisabledString := `
global:
  minio:
    enabled: false
  appConfig:
    object_store:
      enabled: true
      connection:
        secret: gitlab-external-storage
        key: connection
gitlab:
  toolbox:
    backups:
      objectStorage:
        config:
          secret: s3cmd-config
          key: config
`

	_ = chartValuesDisabled.AddFromYAML(chartValuesDisabledString)

	tests := map[string]struct {
		chartValues resource.Values
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
			enabled := MinioEnabled(adapter)

			It("Should render the template", func() {
				Expect(err).To(BeNil())
				Expect(template).NotTo(BeNil())
			})

			It(fmt.Sprintf("Should have %s internal MinIO", name), func() {
				Expect(enabled).To(Equal(test.expected))
			})
		})
	}
})
