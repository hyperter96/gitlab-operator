package gitlab

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("InstallationType", func() {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ConfigMap",
		},
		Data: map[string]string{
			"installation_type": "gitlab-helm-chart",
			"foo":               "bar",
		},
	}

	setInstallationType(cm)

	It("Should change the installation type when configured", func() {
		Expect(cm.Data["installation_type"]).To(Equal(installationType))
	})

	It("Should leave other data fields alone", func() {
		Expect(cm.Data["foo"]).To(Equal("bar"))
	})
})
