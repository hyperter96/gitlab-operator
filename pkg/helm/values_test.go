package helm_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/pkg/helm"
)

var _ = Describe("Values", func() {

	When("Initialized", func() {

		subject := helm.EmptyValues()

		It("Must be empty", func() {
			Expect(subject.AsMap()).To(BeEmpty())
		})

	})

	When("Values added as key-value assignments", func() {

		It("Must store nested values", func() {
			subject := helm.EmptyValues()

			Expect(subject.AddValue("foo.bar", "FOOBAR")).To(BeNil())

			foo, ok := (subject.AsMap()["foo"]).(map[string]interface{})
			Expect(ok).To(BeTrue())
			Expect(foo["bar"]).To(Equal("FOOBAR"))
		})

		It("Must store indexed values", func() {
			subject := helm.EmptyValues()

			Expect(subject.AddValue("foo", "{FOO-0}")).To(BeNil())

			foo, ok := (subject.AsMap()["foo"]).([]interface{})
			Expect(ok).To(BeTrue())
			Expect(foo[0]).To(Equal("FOO-0"))
		})

		It("Must merge and override values", func() {
			subject := helm.EmptyValues()

			Expect(subject.AddValue("foo.bar", "FOOBAR")).To(BeNil())
			Expect(subject.AddValue("foo.baz", "FOOBAZ")).To(BeNil())

			foo, ok := (subject.AsMap()["foo"]).(map[string]interface{})
			Expect(ok).To(BeTrue())
			Expect(foo["bar"]).To(Equal("FOOBAR"))
			Expect(foo["baz"]).To(Equal("FOOBAZ"))

			Expect(subject.AddValue("foo.bar", "{FOOBAR-0}")).To(BeNil())

			fooBar, ok := (foo["bar"]).([]interface{})
			Expect(ok).To(BeTrue())
			Expect(fooBar[0]).To(Equal("FOOBAR-0"))
		})
	})

	When("Values added from file", func() {

		It("Must load file content", func() {
			subject := helm.EmptyValues()

			Expect(subject.AddFromFile("testdata/values.yaml")).To(BeNil())

			Expect(subject.AsMap()["bar"]).To(Equal("BAR"))
			Expect(subject.AsMap()["baz"]).To(Equal("BAZ"))

			foo, ok := (subject.AsMap()["foo"]).(map[string]interface{})
			Expect(ok).To(BeTrue())
			Expect(foo["bar"]).To(Equal("FOOBAR"))

			fooBaz, ok := (foo["baz"]).([]interface{})
			Expect(ok).To(BeTrue())
			Expect(fooBaz[0]).To(Equal("FOOBAZ-0"))
		})

		It("Must merge and override values added later", func() {
			subject := helm.EmptyValues()

			Expect(subject.AddFromFile("testdata/values.yaml")).To(BeNil())
			Expect(subject.AddValue("foo.baz", "FOOBAZ")).To(BeNil())

			foo, ok := (subject.AsMap()["foo"]).(map[string]interface{})
			Expect(ok).To(BeTrue())
			Expect(foo["baz"]).To(Equal("FOOBAZ"))
		})
	})
})
