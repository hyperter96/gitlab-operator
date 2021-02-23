package gitlab_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/gitlab"
)

var _ = Describe("Redis replacement", func() {

	mockCR := GitLabMock()
	adapter := gitlab.NewCustomResourceAdapter(mockCR)

	When("replacing StatefulSet", func() {
		templated := gitlab.RedisStatefulSet(adapter)
		generated := gitlab.RedisStatefulSetDEPRECATED(mockCR)

		It("must completely satisfy the generator function", func() {
			Expect(templated).To(SatisfyReplacement(generated))
		})
	})

	When("replacing ConfigMaps", func() {
		// Only compare the ConfigMaps that the Operator renders.
		// - [x] test-redis-health
		// - [x] test-redis
		// - [ ] test-redis-scripts (unique to Chart)
		templated := gitlab.RedisConfigMaps(adapter)
		templatedCfgMap := gitlab.CfgMapFromList(fmt.Sprintf("%s-redis", mockCR.Name), templated)
		templatedCfgMapScripts := gitlab.CfgMapFromList(fmt.Sprintf("%s-redis-health", mockCR.Name), templated)

		generatedCfgMap := gitlab.RedisConfigMapDEPRECATED(mockCR)              // equal to test-redis from Chart
		generatedCfgMapScripts := gitlab.RedisSciptsConfigMapDEPRECATED(mockCR) // equal to test-redis-health from Chart

		It("must return two ConfigMaps with similar ObjectMeta", func() {
			Expect(templatedCfgMap.ObjectMeta).To(
				SatisfyReplacement(generatedCfgMap.ObjectMeta))
			Expect(templatedCfgMapScripts.ObjectMeta).To(
				SatisfyReplacement(generatedCfgMapScripts.ObjectMeta))
		})

		It("must return two ConfigMaps that contain the same Data items", func() {
			Expect(templatedCfgMap.Data).To(
				SatisfyReplacement(generatedCfgMap.Data))
			Expect(templatedCfgMapScripts.Data).To(
				SatisfyReplacement(generatedCfgMapScripts.Data))
		})
	})

	When("replacing Services", func() {
		// Only compare the services that the Operator renders.
		// - [x] test-redis-headless
		// - [x] test-redis-master
		// - [ ] test-redis-metrics (unique to Chart)
		templated := gitlab.RedisServices(adapter)
		templatedSvc := gitlab.SvcFromList(fmt.Sprintf("%s-redis-master", mockCR.Name), templated)
		templatedSvcHeadless := gitlab.SvcFromList(fmt.Sprintf("%s-redis-headless", mockCR.Name), templated)

		generatedSvc := gitlab.RedisServiceDEPRECATED(mockCR)                 // equal to test-redis-master from Chart
		generatedHeadlessSvc := gitlab.RedisHeadlessServiceDEPRECATED(mockCR) // equal to test-redis-headless from Chart

		It("must completely satisfy the generator function", func() {
			Expect(templatedSvc).To(SatisfyReplacement(generatedSvc))
			Expect(templatedSvcHeadless).To(SatisfyReplacement(generatedHeadlessSvc))
		})
	})

})
