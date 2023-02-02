package objects

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Editor", func() {
	Describe("SetNamespace", func() {
		It("should change the namespace", func() {
			o := setQualifiedName(newObject("foo.bar/v1", "Test"), "test-1", "tests")

			Expect(SetNamespace("default")(o)).To(Succeed())
			Expect(o.GetNamespace()).To(Equal("default"))
		})
	})

	Describe("SetAnnotations", func() {
		It("should merge the annotations", func() {
			o := setAnnotations(newObject("foo.bar/v1", "Test"), map[string]string{
				"foo": "foo",
				"bar": "bar-1",
			})

			Expect(SetAnnotations(map[string]string{
				"bar": "bar-2",
				"baz": "baz",
			})(o)).To(Succeed())
			Expect(o.GetAnnotations()).To(Equal(map[string]string{
				"foo": "foo",
				"bar": "bar-2",
				"baz": "baz",
			}))
		})
	})
})
