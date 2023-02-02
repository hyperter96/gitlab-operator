package internal

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Internal helpers", func() {
	Context("Truncating strings", func() {
		testString := "abcdefghijklmnopqrstuvwxyz"
		testStringLen := len(testString)

		testCases := []struct {
			when           string
			it             string
			length         int
			expectedResult string
		}{
			{"Passing a number equal to the string length", "Should return the input string", testStringLen, testString},
			{"Passing a number greater than the string length", "Should return the input string", 50, testString},
			{"Passing a number greater than zero and less than the string length", "Should return the input string truncated", 5, "abcde"},
		}

		for _, tc := range testCases {
			When(tc.when, func() {
				It(tc.it, func() {
					result, err := Truncate(testString, tc.length)
					Expect(err).To(BeNil())
					Expect(result).To(Equal(tc.expectedResult))
				})
			})
		}

		When("Passing a negative number", func() {
			It("Returns an error", func() {
				result, err := Truncate(testString, -1)
				Expect(err).NotTo(BeNil())
				Expect(result).To(Equal(""))
			})
		})
	})
})
