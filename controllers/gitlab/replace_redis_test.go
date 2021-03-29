package gitlab

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/helpers"
)

var _ = Describe("Redis replacement", func() {

	mockCR := GitLabMock()
	adapter := helpers.NewCustomResourceAdapter(mockCR)

	When("replacing StatefulSet", func() {
		templated := RedisStatefulSet(adapter)
		generated := RedisStatefulSetDEPRECATED(mockCR)

		It("must completely satisfy the generator function", func() {
			Expect(templated).To(SatisfyReplacement(generated))
		})
	})

	When("replacing ConfigMaps", func() {
		It("must return two ConfigMaps with similar ObjectMeta", func() {
			// Only compare the ConfigMaps that the Operator renders.
			// - [x] test-redis-health
			// - [x] test-redis
			// - [ ] test-redis-scripts (unique to Chart)
			templated := RedisConfigMaps(adapter)
			templatedCfgMap := CfgMapFromList(fmt.Sprintf("%s-redis", mockCR.Name), templated)
			templatedCfgMapScripts := CfgMapFromList(fmt.Sprintf("%s-redis-health", mockCR.Name), templated)

			generatedCfgMap := RedisConfigMapDEPRECATED(mockCR)              // equal to test-redis from Chart
			generatedCfgMapScripts := RedisSciptsConfigMapDEPRECATED(mockCR) // equal to test-redis-health from Chart

			Expect(templatedCfgMap.ObjectMeta).To(
				SatisfyReplacement(generatedCfgMap.ObjectMeta))
			Expect(templatedCfgMapScripts.ObjectMeta).To(
				SatisfyReplacement(generatedCfgMapScripts.ObjectMeta))
		})

		It("must return two ConfigMaps that contain the same Data items", func() {
			// Only compare the ConfigMaps that the Operator renders.
			// - [x] test-redis-health
			// - [x] test-redis
			// - [ ] test-redis-scripts (unique to Chart)
			templated := RedisConfigMaps(adapter)
			templatedCfgMap := CfgMapFromList(fmt.Sprintf("%s-redis", mockCR.Name), templated)
			templatedCfgMapScripts := CfgMapFromList(fmt.Sprintf("%s-redis-health", mockCR.Name), templated)

			generatedCfgMap := RedisConfigMapDEPRECATED(mockCR)              // equal to test-redis from Chart
			generatedCfgMapScripts := RedisSciptsConfigMapDEPRECATED(mockCR) // equal to test-redis-health from Chart

			Expect(templatedCfgMap.Data).To(
				SatisfyReplacement(generatedCfgMap.Data))
			Expect(templatedCfgMapScripts.Data).To(
				SatisfyReplacement(generatedCfgMapScripts.Data))
		})
	})

	When("replacing Services", func() {
		It("must completely satisfy the generator function", func() {
			// Only compare the services that the Operator renders.
			// - [x] test-redis-headless
			// - [x] test-redis-master
			// - [ ] test-redis-metrics (unique to Chart)
			templated := RedisServices(adapter)
			templatedSvc := SvcFromList(fmt.Sprintf("%s-redis-master", mockCR.Name), templated)
			templatedSvcHeadless := SvcFromList(fmt.Sprintf("%s-redis-headless", mockCR.Name), templated)

			generatedSvc := RedisServiceDEPRECATED(mockCR)                 // equal to test-redis-master from Chart
			generatedHeadlessSvc := RedisHeadlessServiceDEPRECATED(mockCR) // equal to test-redis-headless from Chart

			Expect(templatedSvc).To(SatisfyReplacement(generatedSvc))
			Expect(templatedSvcHeadless).To(SatisfyReplacement(generatedHeadlessSvc))
		})
	})

})
