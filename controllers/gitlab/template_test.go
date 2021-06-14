package gitlab

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	gitlabv1beta1 "gitlab.com/gitlab-org/cloud-native/gitlab-operator/api/v1beta1"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	chartVersions = AvailableChartVersions()
	chartValues   = helm.EmptyValues()
	namespace     = os.Getenv("HELM_NAMESPACE")
)

var _ = Describe("CustomResourceAdapter", func() {

	if namespace == "" {
		namespace = "default"
	}

	mockGitLab1 := &gitlabv1beta1.GitLab{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: namespace,
		},
		Spec: gitlabv1beta1.GitLabSpec{
			Chart: gitlabv1beta1.GitLabChartSpec{
				Version: chartVersions[0],
			},
		},
	}

	mockGitLab2 := &gitlabv1beta1.GitLab{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: namespace,
		},
		Spec: gitlabv1beta1.GitLabSpec{
			Chart: gitlabv1beta1.GitLabChartSpec{
				Version: chartVersions[1],
			},
		},
	}

	/*
	 * All tests are packed together here to avoid rendering GitLab Chart repeatedly.
	 * This is done to speed up the test.
	 */

	It("must render the template only when the CR has changed", func() {
		adapter1 := NewCustomResourceAdapter(mockGitLab1)
		adapter2 := NewCustomResourceAdapter(mockGitLab2)

		template1, err := GetTemplate(adapter1)

		Expect(err).To(BeNil())
		Expect(template1).NotTo(BeNil())

		template1prime, err := GetTemplate(adapter1)

		Expect(err).To(BeNil())
		Expect(template1prime).To(BeIdenticalTo(template1))

		template2, err := GetTemplate(adapter2)

		Expect(err).To(BeNil())
		Expect(template2).NotTo(BeNil())
		Expect(template2).NotTo(BeIdenticalTo(template1))

		chartInfo1 := template1.Query().
			ObjectByKindAndName("ConfigMap", "test-gitlab-chart-info").(*corev1.ConfigMap)
		chartInfo2 := template2.Query().
			ObjectByKindAndName("ConfigMap", "test-gitlab-chart-info").(*corev1.ConfigMap)

		Expect(chartInfo1).NotTo(BeNil())
		Expect(chartInfo1.Namespace).To(Equal(namespace))
		Expect(chartInfo1.Labels["release"]).To(Equal("test"))
		Expect(chartInfo1.Data["gitlabChartVersion"]).To(Equal(chartVersions[0]))

		Expect(chartInfo2).NotTo(BeNil())
		Expect(chartInfo1.Namespace).To(Equal(namespace))
		Expect(chartInfo1.Labels["release"]).To(Equal("test"))
		Expect(chartInfo2.Data["gitlabChartVersion"]).To(Equal(chartVersions[1]))
	})
})
