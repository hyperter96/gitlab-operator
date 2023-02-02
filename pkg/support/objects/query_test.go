package objects

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Selector", func() {
	Describe("ByKind", func() {
		It("matches different styles of kind", func() {
			o1 := newObject("foo.bar/v1", "Test")
			o2 := newObject("foo.bar/v2", "Test")

			Expect(ByKind("Test")(o1)).To(BeTrue())
			Expect(ByKind("Test.foo.bar")(o1)).To(BeTrue())
			Expect(ByKind("Test.v1.foo.bar")(o1)).To(BeTrue())

			Expect(ByKind("Test")(o2)).To(BeTrue())
			Expect(ByKind("Test.foo.bar")(o2)).To(BeTrue())
			Expect(ByKind("Test.v1.foo.bar")(o2)).To(BeFalse())
			Expect(ByKind("Test.foo.bar.baz")(o2)).To(BeFalse())
		})
	})

	Describe("ByName", func() {
		It("matches name", func() {
			o1 := setQualifiedName(newObject("foo.bar/v1", "Test"), "test-1", "tests")
			o2 := setQualifiedName(newObject("foo.bar/v1", "AnotherTest"), "test-2", "tests")

			Expect(ByName("test-1")(o1)).To(BeTrue())
			Expect(ByName("test-1")(o2)).To(BeFalse())
			Expect(ByName("test-2")(o1)).To(BeFalse())
			Expect(ByName("test-2")(o2)).To(BeTrue())
		})
	})

	Describe("ByNamespace", func() {
		It("matches namespace", func() {
			o1 := setQualifiedName(newObject("foo.bar/v1", "Test"), "test-1", "tests")
			o2 := setQualifiedName(newObject("foo.bar/v2", "Test"), "test-1", "default")

			Expect(ByNamespace("tests")(o1)).To(BeTrue())
			Expect(ByNamespace("tests")(o2)).To(BeFalse())
			Expect(ByNamespace("default")(o1)).To(BeFalse())
			Expect(ByNamespace("default")(o2)).To(BeTrue())
		})
	})

	Describe("ByLabels", func() {
		It("matches labels", func() {
			o1 := setLabels(setQualifiedName(newObject("foo.bar/v1", "Test"), "test-1", "tests"),
				map[string]string{
					"app": "test",
					"foo": "foo-1",
					"bar": "bar",
				})
			o2 := setLabels(setQualifiedName(newObject("foo.bar/v2", "Test"), "test-1", "tests"),
				map[string]string{
					"app": "test",
					"foo": "foo-2",
					"baz": "baz",
				})

			Expect(ByLabels(map[string]string{
				"app": "test",
			})(o1)).To(BeTrue())
			Expect(ByLabels(map[string]string{
				"app": "test",
			})(o2)).To(BeTrue())

			Expect(ByLabels(map[string]string{
				"app": "test",
				"foo": "foo-1",
			})(o1)).To(BeTrue())
			Expect(ByLabels(map[string]string{
				"app": "test",
				"foo": "foo-1",
			})(o2)).To(BeFalse())

			Expect(ByLabels(map[string]string{
				"baz": "baz",
			})(o1)).To(BeFalse())
			Expect(ByLabels(map[string]string{
				"baz": "baz",
			})(o2)).To(BeTrue())

			Expect(ByLabels(map[string]string{
				"app": "test",
				"foo": "foo-1",
				"bar": "bar",
				"baz": "baz",
			})(o1)).To(BeFalse())
		})
	})

	Describe("ByComponent", func() {
		It("matches components", func() {
			o1 := setLabels(setQualifiedName(newObject("foo.bar/v1", "Test"), "test-1", "tests"),
				map[string]string{
					"app": "test",
				})
			o2 := setLabels(setQualifiedName(newObject("foo.bar/v2", "Test"), "test-1", "tests"),
				map[string]string{
					"app.kubernetes.io/component": "test",
				})

			Expect(ByComponent("test")(o1)).To(BeTrue())
			Expect(ByComponent("test")(o2)).To(BeTrue())
			Expect(ByComponent("foo")(o1)).To(BeFalse())
			Expect(ByComponent("bar")(o2)).To(BeFalse())
		})
	})

	Describe("All", func() {
		It("passes when all match", func() {
			o := setQualifiedName(newObject("foo.bar/v1", "Test"), "test-1", "tests")

			Expect(All(ByKind("Test"), ByName("test-1"))(o)).To(BeTrue())
			Expect(All(ByKind("Test"), ByName("test-1"), ByNamespace("tests"))(o)).To(BeTrue())
		})

		It("fails when one does not match", func() {
			o := setQualifiedName(newObject("foo.bar/v1", "Test"), "test-1", "tests")

			Expect(All(ByKind("Test"), ByName("another-test"))(o)).To(BeFalse())
			Expect(All(ByKind("Test"), ByName("test-1"), ByNamespace("default"))(o)).To(BeFalse())
		})
	})

	Describe("Any", func() {
		It("passes when one matches", func() {
			o := setQualifiedName(newObject("foo.bar/v1", "Test"), "test-1", "tests")

			Expect(Any(ByKind("Test"), ByName("another-test"))(o)).To(BeTrue())
			Expect(Any(ByKind("Test"), ByName("test-1"), ByNamespace("default"))(o)).To(BeTrue())
		})

		It("fails when none matches", func() {
			o := setQualifiedName(newObject("foo.bar/v1", "Test"), "test-1", "tests")

			Expect(Any(ByKind("AnotherTest"), ByName("another-test"))(o)).To(BeFalse())
			Expect(Any(ByKind("AnotherTest"), ByName("another-test"), ByNamespace("default"))(o)).To(BeFalse())
		})
	})

	Describe("None", func() {
		It("passes when none matches", func() {
			o := setQualifiedName(newObject("foo.bar/v1", "Test"), "test-1", "tests")

			Expect(None(ByKind("AnotherTest"), ByName("another-test"))(o)).To(BeTrue())
			Expect(None(ByKind("AnotherTest"), ByName("another-test"), ByNamespace("default"))(o)).To(BeTrue())
		})

		It("fails when one one matches", func() {
			o := setQualifiedName(newObject("foo.bar/v1", "Test"), "test-1", "tests")

			Expect(All(ByKind("Test"), ByName("another-test"))(o)).To(BeFalse())
			Expect(All(ByKind("Test"), ByName("another-test"), ByNamespace("default"))(o)).To(BeFalse())
		})
	})
})
