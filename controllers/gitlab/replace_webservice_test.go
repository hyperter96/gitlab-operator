package gitlab

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/helpers"
)

var _ = Describe("Webservice replacement", func() {

	mockCR := GitLabMock()
	adapter := helpers.NewCustomResourceAdapter(mockCR)

	When("replacing Deployment", func() {
		templated := WebserviceDeployment(adapter)
		generated := WebserviceDeploymentDEPRECATED(mockCR)

		It("must completely satisfy the generator function", func() {
			Expect(templated).To(
				SatisfyReplacement(generated))
		})
	})

	When("replacing ConfigMap", func() {
		templated := WebserviceConfigMaps(adapter)
		generated := WebserviceConfigMapDEPRECATED(mockCR)

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
		templated := WebserviceService(adapter)
		generated := WebserviceServiceDEPRECATED(mockCR)

		It("must completely satisfy the generator function", func() {
			Expect(templated).To(
				SatisfyReplacement(generated))
		})
	})

	When("replacing Ingress", func() {
		templated := WebserviceIngress(adapter)
		generated := IngressDEPRECATED(adapter)

		It("must completely satisfy the generator function", func() {
			Expect(templated).To(
				SatisfyReplacement(generated))
		})
	})
})
