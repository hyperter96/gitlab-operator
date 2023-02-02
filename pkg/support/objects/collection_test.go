package objects

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Collection", func() {
	It("is empty after initialization", func() {
		Expect(Collection{}).To(BeEmpty())
	})

	It("is a slice", func() {
		c := Collection{
			newObject("foo.bar/v1", "Test"),
		}

		Expect(c).To(HaveLen(1))
		Expect(c[0].GetObjectKind().GroupVersionKind().String()).To(Equal("foo.bar/v1, Kind=Test"))
	})

	Describe("Append", func() {
		It("adds new object without changing the reference", func() {
			c := Collection{}
			cRef := &c

			c.Append(newObject("foo.bar/v1", "Test1"))

			Expect(c).To(HaveLen(1))
			Expect(&c).To(BeIdenticalTo(cRef))
		})

		It("grows the collection", func() {
			c := Collection{}

			Expect(c.Empty()).To(BeTrue())

			c.Append(newObject("foo.bar/v1", "Test1"))
			c.Append(newObject("foo.bar/v1", "Test2"))
			c.Append(newObject("foo.bar/v1", "Test3"))

			Expect(c.Empty()).To(BeFalse())
			Expect(c).To(HaveLen(3))
		})
	})

	Describe("First", func() {
		It("returns nil when empty", func() {
			c := Collection{}

			Expect(c.First()).To(BeNil())
		})

		It("returns the first item of the collection", func() {
			c := Collection{
				newObject("foo.bar/v1", "Test1"),
				newObject("foo.bar/v1", "Test2"),
			}

			Expect(c.First().GetObjectKind().GroupVersionKind().Kind).To(Equal("Test1"))
		})
	})

	Describe("Query", func() {
		falseSelector := func(client.Object) bool { return false }
		trueSelector := func(client.Object) bool { return true }

		It("returns an empty collection when none of the objects match", func() {
			c := Collection{
				newObject("foo.bar/v1", "Test1"),
				newObject("foo.bar/v1", "Test2"),
			}

			Expect(c.Empty()).To(BeFalse())
			Expect(c.Query(falseSelector).Empty()).To(BeTrue())
			Expect(c.Query(trueSelector, falseSelector).Empty()).To(BeTrue())
		})

		It("returns an empty collection when there is no selector", func() {
			c := Collection{
				newObject("foo.bar/v1", "Test1"),
				newObject("foo.bar/v1", "Test2"),
			}

			q := c.Query()

			Expect(q).To(BeEmpty())
			Expect(q).NotTo(BeIdenticalTo(c))
		})

		It("returns a new collection with the same objects that match", func() {
			c := Collection{
				newObject("foo.bar/v1", "Test1"),
				newObject("foo.bar/v1", "Test2"),
			}

			q := c.Query(ByKind("Test1"))

			Expect(q).To(HaveLen(1))
			Expect(q).NotTo(BeIdenticalTo(c))
			Expect(q[0]).To(BeIdenticalTo(c[0]))
		})
	})

	Describe("Edit", func() {
		setName := func(o client.Object) error {
			o.SetName("test-" + o.GetObjectKind().GroupVersionKind().Kind)

			return nil
		}

		rejectTest2 := func(o client.Object) error {
			if o.GetObjectKind().GroupVersionKind().Kind == "Test2" {
				return fmt.Errorf("don't like it")
			}

			return nil
		}

		ignoreTest2 := func(o client.Object) error {
			if o.GetObjectKind().GroupVersionKind().Kind == "Test2" {
				return NewTypeMismatchError("Test1", "Test2")
			}

			return nil
		}

		It("changes all objects without changing their reference", func() {
			c := Collection{
				newObject("foo.bar/v1", "Test1"),
				newObject("foo.bar/v1", "Test2"),
			}

			c0Ref := c[0]
			c1Ref := c[1]

			count, err := c.Edit(SetNamespace("test"))

			Expect(err).To(Succeed())
			Expect(count).To(Equal(2))

			Expect(c[0].GetNamespace()).To(Equal("test"))
			Expect(c[1].GetNamespace()).To(Equal("test"))

			Expect(c0Ref).To(BeIdenticalTo(c[0]))
			Expect(c1Ref).To(BeIdenticalTo(c[1]))
		})

		It("accumulates changes from all editors to objects", func() {
			c := Collection{
				newObject("foo.bar/v1", "Test1"),
				newObject("foo.bar/v1", "Test2"),
			}

			count, err := c.Edit(setName, SetNamespace("test"))

			Expect(err).To(Succeed())
			Expect(count).To(Equal(2))

			Expect(c[0].GetNamespace()).To(Equal("test"))
			Expect(c[1].GetNamespace()).To(Equal("test"))

			Expect(c[0].GetName()).To(Equal("test-Test1"))
			Expect(c[1].GetName()).To(Equal("test-Test2"))
		})

		It("fails half-way through when encounters an error", func() {
			c := Collection{
				newObject("foo.bar/v1", "Test1"),
				newObject("foo.bar/v1", "Test2"),
				newObject("foo.bar/v1", "Test3"),
			}

			count, err := c.Edit(SetNamespace("test"), rejectTest2, setName)

			Expect(err).NotTo(Succeed())
			Expect(count).To(Equal(1))

			Expect(c[0].GetNamespace()).To(Equal("test"))
			Expect(c[1].GetNamespace()).To(Equal("test"))

			Expect(c[0].GetName()).To(Equal("test-Test1"))
			Expect(c[1].GetName()).To(Equal(""))
		})

		It("ignores type mismatch error and continues to change other objects", func() {
			c := Collection{
				newObject("foo.bar/v1", "Test1"),
				newObject("foo.bar/v1", "Test2"),
				newObject("foo.bar/v1", "Test3"),
			}

			count, err := c.Edit(SetNamespace("test"), ignoreTest2, setName)

			Expect(err).To(Succeed())
			Expect(count).To(Equal(3))

			Expect(c[0].GetNamespace()).To(Equal("test"))
			Expect(c[1].GetNamespace()).To(Equal("test"))
			Expect(c[2].GetNamespace()).To(Equal("test"))

			Expect(c[0].GetName()).To(Equal("test-Test1"))
			Expect(c[1].GetName()).To(Equal("test-Test2"))
			Expect(c[2].GetName()).To(Equal("test-Test3"))
		})
	})

	C1 := Collection{
		setQualifiedName(newObject("foo.bar/v1", "Test1"), "test1", "test"),
		setQualifiedName(newObject("foo.bar/v1", "Test2"), "test2", "test"),
	}
	C2 := Collection{
		setQualifiedName(newObject("foo.bar/v1", "Test1"), "test1", "default"),
		setQualifiedName(newObject("foo.bar/v1", "Test2"), "test2", "default"),
	}
	C3 := Collection{
		setQualifiedName(newObject("foo.bar/v1", "Test1"), "test1", "test"),
		setQualifiedName(newObject("foo.bar/v1", "Test2"), "test2", "default"),
	}
	R1 := Collection{
		setAnnotations(setQualifiedName(newObject("foo.bar/v1", "Test1"), "test1", "test"),
			map[string]string{"index": "1"}),
		setQualifiedName(newObject("foo.bar/v1", "Test2"), "test2", "test"),
		setAnnotations(setQualifiedName(newObject("foo.bar/v1", "Test1"), "test1", "test"),
			map[string]string{"index": "2"}),
	}
	R2 := Collection{
		setQualifiedName(newObject("foo.bar/v1", "Test1"), "test1", "test"),
		setAnnotations(setQualifiedName(newObject("foo.bar/v1", "Test2"), "test2", "test"),
			map[string]string{"index": "1"}),
		setAnnotations(setQualifiedName(newObject("foo.bar/v1", "Test2"), "test2", "test"),
			map[string]string{"index": "2"}),
	}

	C1C3Intersection := Collection{
		setQualifiedName(newObject("foo.bar/v1", "Test1"), "test1", "test"),
	}
	C1C2Difference := C1.Clone()
	C1C3Difference := Collection{
		setQualifiedName(newObject("foo.bar/v1", "Test2"), "test2", "test"),
	}
	C3C1Difference := Collection{
		setQualifiedName(newObject("foo.bar/v1", "Test2"), "test2", "default"),
	}
	R1C3Intersection := Collection{
		setAnnotations(setQualifiedName(newObject("foo.bar/v1", "Test1"), "test1", "test"),
			map[string]string{"index": "1"}),
		setAnnotations(setQualifiedName(newObject("foo.bar/v1", "Test1"), "test1", "test"),
			map[string]string{"index": "2"}),
	}
	R2C3Difference := Collection{
		setAnnotations(setQualifiedName(newObject("foo.bar/v1", "Test2"), "test2", "test"),
			map[string]string{"index": "1"}),
		setAnnotations(setQualifiedName(newObject("foo.bar/v1", "Test2"), "test2", "test"),
			map[string]string{"index": "2"}),
	}

	Describe("Intersection", func() {
		It("returns an empty collection when there is no intersection", func() {
			Expect(C1.Intersection(Collection{})).To(BeEmpty())
			Expect(C1.Intersection(C2)).To(BeEmpty())
		})

		It("returns the intersecting objects", func() {
			Expect(C1.Intersection(C3)).To(Equal(C1C3Intersection))
		})

		It("includes duplicate objects", func() {
			Expect(R1.Intersection(C3)).To(Equal(R1C3Intersection))
		})
	})

	Describe("Difference", func() {
		It("returns an empty collection when there is no difference", func() {
			Expect(C1.Difference(C1.Clone())).To(BeEmpty())
			Expect(Collection{}.Difference(C1)).To(BeEmpty())
		})

		It("returns the LHS when the one is null", func() {
			Expect(C1.Difference(Collection{})).To(Equal(C1))
		})

		It("returns the differing objects", func() {
			Expect(C1.Difference(C2)).To(Equal(C1C2Difference))
			Expect(C1.Difference(C3)).To(Equal(C1C3Difference))
			Expect(C3.Difference(C1)).To(Equal(C3C1Difference))
		})

		It("includes duplicate objects", func() {
			Expect(R2.Difference(C3)).To(Equal(R2C3Difference))
		})
	})

	Describe("Clone", func() {
		It("creates a deep copy of the collection", func() {
			c := Collection{
				newObject("foo.bar/v1", "Test1"),
				newObject("foo.bar/v1", "Test2"),
			}

			cc := c.Clone()

			Expect(cc).To(Equal(c))

			Expect(cc).NotTo(BeIdenticalTo(c))
			Expect(cc[0]).NotTo(BeIdenticalTo(c[0]))
			Expect(cc[1]).NotTo(BeIdenticalTo(c[1]))
		})
	})
})
