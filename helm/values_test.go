package helm

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Values", func() {

	When("initialized", func() {

		subject := EmptyValues()

		It("must be empty", func() {
			Expect(subject.AsMap()).To(BeEmpty())
		})

	})

	When("values added as key-value assignments", func() {

		It("must store nested values", func() {
			subject := EmptyValues()

			Expect(subject.AddValue("foo.bar", "FOOBAR")).To(BeNil())

			foo, ok := (subject.AsMap()["foo"]).(map[string]interface{})
			Expect(ok).To(BeTrue())
			Expect(foo["bar"]).To(Equal("FOOBAR"))
		})

		It("must store indexed values", func() {
			subject := EmptyValues()

			Expect(subject.AddValue("foo", "{FOO-0}")).To(BeNil())

			foo, ok := (subject.AsMap()["foo"]).([]interface{})
			Expect(ok).To(BeTrue())
			Expect(foo[0]).To(Equal("FOO-0"))
		})

		It("must merge and override values", func() {
			subject := EmptyValues()

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

	When("values added from file", func() {

		It("must load file content", func() {
			subject := EmptyValues()

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

		It("must merge and override values added later", func() {
			subject := EmptyValues()

			Expect(subject.AddFromFile("testdata/values.yaml")).To(BeNil())
			Expect(subject.AddValue("foo.baz", "FOOBAZ")).To(BeNil())

			foo, ok := (subject.AsMap()["foo"]).(map[string]interface{})
			Expect(ok).To(BeTrue())
			Expect(foo["baz"]).To(Equal("FOOBAZ"))
		})
	})

	Context("getter", func() {

		subject := EmptyValues()

		Expect(subject.AddValue("foo.bar", "FOOBAR")).To(BeNil())

		It("must return the root node when key is empty", func() {
			value, err := subject.GetValue("")
			Expect(err).To(Succeed())
			Expect(value).To(Equal(map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": "FOOBAR",
				},
			}))
		})

		It("must read leaf node values and return nil when missing", func() {
			value, err := subject.GetValue("foo.bar")
			Expect(err).To(Succeed())
			Expect(value).To(Equal("FOOBAR"))

			value, err = subject.GetValue("foo.baz")
			Expect(err).To(Succeed())
			Expect(value).To(BeNil())
		})

		It("must return error when key element is missing or is a leaf node", func() {
			value, err := subject.GetValue("foo.baz.bar")
			Expect(err).To(HaveOccurred())
			Expect(value).To(BeNil())

			value, err = subject.GetValue("foo.bar.baz")
			Expect(err).To(HaveOccurred())
			Expect(value).To(BeNil())
		})
	})

	Context("setter", func() {

		It("must fail when key is empty", func() {
			Expect(EmptyValues().SetValue("", "Oops!")).To(HaveOccurred())
		})

		It("must create the full path even when it is missing", func() {
			subject := EmptyValues()

			Expect(subject.SetValue("baz", "BAZ")).To(Succeed())
			Expect(subject.SetValue("foo.bar", map[string]interface{}{
				"box": "BOX",
				"baz": []int{1, 2, 3},
			})).To(Succeed())

			Expect(subject.AsMap()).To(Equal(map[string]interface{}{
				"baz": "BAZ",
				"foo": map[string]interface{}{
					"bar": map[string]interface{}{
						"box": "BOX",
						"baz": []int{1, 2, 3},
					},
				},
			}))
		})

		It("must return error when key element is a leaf node", func() {
			subject := EmptyValues()

			Expect(subject.SetValue("foo.bar", "FOOBAR")).To(Succeed())
			Expect(subject.SetValue("foo.bar.baz", "Oops!")).To(HaveOccurred())

			Expect(subject.AsMap()).To(Equal(map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": "FOOBAR",
				},
			}))
		})

	})
})
