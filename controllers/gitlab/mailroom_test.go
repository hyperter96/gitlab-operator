package gitlab

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
)

var _ = Describe("CustomResourceAdapter", func() {

	if namespace == "" {
		namespace = "default"
	}

	Context("Mailroom", func() {
		When("Mailroom is enabled", func() {
			chartValues := helm.EmptyValues()
			_ = chartValues.SetValue(GitLabMailroomEnabled, true)
			_ = chartValues.SetValue(IncomingEmailEnabled, true)

			adapter := createMockAdapter(namespace, chartVersions[0], chartValues)
			template, err := GetTemplate(adapter)

			//dumpTemplate(adapter)

			enabled := MailroomEnabled(adapter)
			configMap := MailroomConfigMap(adapter)
			hpa := MailroomHPA(adapter)
			deployment := MailroomDeployment(adapter)
			networkPolicy := MailroomNetworkPolicy(adapter)
			serviceAccount := MailroomServiceAccount(adapter)

			It("Should render the template", func() {
				Expect(err).To(BeNil())
				Expect(template).NotTo(BeNil())
			})

			It("Should contain Mailroom resources", func() {
				Expect(enabled).To(BeTrue())
				Expect(deployment).NotTo(BeNil())
				Expect(configMap).NotTo(BeNil())
				Expect(hpa).To(BeNil())
				Expect(networkPolicy).To(BeNil())
				Expect(serviceAccount).NotTo(BeNil())
			})

		})

		When("Mailroom is disabled", func() {
			chartValues := helm.EmptyValues()
			_ = chartValues.SetValue(GitLabMailroomEnabled, false)

			adapter := createMockAdapter(namespace, chartVersions[0], chartValues)
			template, err := GetTemplate(adapter)

			enabled := MailroomEnabled(adapter)
			configMap := MailroomConfigMap(adapter)
			hpa := MailroomHPA(adapter)
			deployment := MailroomDeployment(adapter)
			networkPolicy := MailroomNetworkPolicy(adapter)
			serviceAccount := MailroomServiceAccount(adapter)

			It("Should render the template", func() {
				Expect(err).To(BeNil())
				Expect(template).NotTo(BeNil())
			})

			It("Should not contain Mailroom resources", func() {
				Expect(enabled).To(BeFalse())
				Expect(deployment).To(BeNil())
				Expect(configMap).To(BeNil())
				Expect(hpa).To(BeNil())
				Expect(networkPolicy).To(BeNil())
				Expect(serviceAccount).To(BeNil())
			})

		})
	})
})
