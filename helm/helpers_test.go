package helm

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Template", func() {
	It("must return all objects when the selector matches all", func() {
		template, err := loadTemplate()
		Expect(err).To(BeNil())

		selectedObjects, err := template.GetObjects(TrueSelector)
		Expect(err).To(BeNil())
		Expect(selectedObjects).To(Equal(template.Objects()))
	})

	It("must delete no object when the selector does not match any", func() {
		template, err := loadTemplate()
		Expect(err).To(BeNil())

		deletedCount, err := template.DeleteObjects(FalseSelector)
		Expect(err).To(BeNil())
		Expect(deletedCount).To(BeZero())
	})
})
