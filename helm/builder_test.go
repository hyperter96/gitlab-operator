package helm_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/helm"
)

var _ = Describe("Builder", func() {

	builder := helm.NewBuilder("foo")
	namespace := os.Getenv("HELM_NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}

	It("must be empty and use default settings", func() {
		Expect(builder.Chart()).To(Equal("foo"))
		Expect(builder.Namespace()).To(Equal(namespace))
		Expect(builder.ReleaseName()).To(Equal("ephemeral"))
		Expect(builder.HooksDisabled()).To(BeFalse())
	})

})
