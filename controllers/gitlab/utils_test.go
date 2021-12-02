package gitlab

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Utils", func() {

	When("Chart Version is less than 5.5.0", func() {
		It("ToolboxComponentName should return task-runner", func() {
			Expect(ToolboxComponentName("5.4.2")).To(Equal("task-runner"))
		})
	})

	When("Chart Version is equal to 5.5.0", func() {
		It("ToolboxComponentName should return toolbox", func() {
			Expect(ToolboxComponentName("5.5.0")).To(Equal("toolbox"))
		})
	})

	When("Chart Version is greater than 5.5.0", func() {
		It("ToolboxComponentName should return toolbox", func() {
			Expect(ToolboxComponentName("5.6.0")).To(Equal("toolbox"))
		})
	})

})
