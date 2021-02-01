package gitlab_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/gitlab"
)

const (
	// WebserviceComponentName is the common name of Webservice.
	WebserviceComponentName = "webservice"
)

var _ = Describe("Webservice replacement", func() {

	mockCR := GitLabMock()
	adapter := gitlab.NewCustomResourceAdapter(mockCR)

	When("replacing Deployment", func() {
		templated := gitlab.WebserviceDeployment(adapter)
		generated := gitlab.WebserviceDeploymentDEPRECATED(mockCR)

		It("must completely satisfy the generator function", func() {
			Expect(templated).To(
				SatisfyReplacement(generated))
		})
	})

	When("replacing ConfigMap", func() {
		templated := gitlab.WebserviceConfigMaps(adapter)
		generated := gitlab.WebserviceConfigMapDEPRECATED(mockCR)

		It("must return two ConfigMaps with similar ObjectMeta", func() {
			Expect(templated).To(HaveLen(2))
			Expect(templated[0].ObjectMeta).To(
				SatisfyReplacement(generated.ObjectMeta))
			Expect(templated[1].ObjectMeta).To(
				SatisfyReplacement(generated.ObjectMeta))
		})

		It("must return two ConfigMaps that with similar Data items", func() {
			Expect(templated).To(HaveLen(2))

			generatedData := map[string]string{}
			templatedData := map[string]string{}

			for k, v := range generated.Data {
				generatedData[k] = v
			}

			for _, cfgMap := range templated {
				for k, v := range cfgMap.Data {
					templatedData[k] = v
				}
			}

			Expect(templatedData).To(SatisfyReplacement(generatedData))

		})
	})

	When("replacing Service", func() {
		templated := gitlab.WebserviceService(adapter)
		generated := gitlab.WebserviceServiceDEPRECATED(mockCR)

		It("must completely satisfy the generator function", func() {
			Expect(templated).To(
				SatisfyReplacement(generated))
		})
	})

})
