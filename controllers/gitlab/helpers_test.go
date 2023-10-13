package gitlab

import (
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

func matchAllElements(match OmegaMatcher, components ...string) Elements {
	elements := Elements{}

	for _, c := range components {
		elements[c] = match
	}

	return elements
}
