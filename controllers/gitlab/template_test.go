package gitlab

import (
	"encoding/json"
	"os"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	gitlabv1beta1 "gitlab.com/gitlab-org/cloud-native/gitlab-operator/api/v1beta1"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"

	k8sjson "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes/scheme"
)

var (
	chartVersions []string = AvailableChartVersions()
	namespace     string   = os.Getenv("HELM_NAMESPACE")
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

	Context("GitLab Pages", func() {
		When("Pages is enabled", func() {
			chartValues := helm.EmptyValues()
			_ = chartValues.SetValue(globalPagesEnabled, true)

			mockGitLab := &gitlabv1beta1.GitLab{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: namespace,
				},
				Spec: gitlabv1beta1.GitLabSpec{
					Chart: gitlabv1beta1.GitLabChartSpec{
						Version: chartVersions[0],
						Values: gitlabv1beta1.ChartValues{
							Object: chartValues.AsMap(),
						},
					},
				},
			}

			adapter := NewCustomResourceAdapter(mockGitLab)
			template, err := GetTemplate(adapter)

			enabled := PagesEnabled(adapter)
			configMap := PagesConfigMap(adapter)
			service := PagesService(adapter)
			deployment := PagesDeployment(adapter)
			ingress := PagesIngress(adapter)

			It("Should render the template", func() {
				Expect(err).To(BeNil())
				Expect(template).NotTo(BeNil())
			})

			It("Should contain Pages resources", func() {
				Expect(enabled).To(BeTrue())
				Expect(configMap).NotTo(BeNil())
				Expect(service).NotTo(BeNil())
				Expect(deployment).NotTo(BeNil())
				Expect(ingress).NotTo(BeNil())
			})
		})

		When("Pages is disabled", func() {
			chartValues := helm.EmptyValues()
			_ = chartValues.SetValue(globalPagesEnabled, false)

			mockGitLab := &gitlabv1beta1.GitLab{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: namespace,
				},
				Spec: gitlabv1beta1.GitLabSpec{
					Chart: gitlabv1beta1.GitLabChartSpec{
						Version: chartVersions[0],
						Values: gitlabv1beta1.ChartValues{
							Object: chartValues.AsMap(),
						},
					},
				},
			}

			adapter := NewCustomResourceAdapter(mockGitLab)
			template, err := GetTemplate(adapter)

			enabled := PagesEnabled(adapter)
			configMap := PagesConfigMap(adapter)
			service := PagesService(adapter)
			deployment := PagesDeployment(adapter)
			ingress := PagesIngress(adapter)

			It("Should render the template", func() {
				Expect(err).To(BeNil())
				Expect(template).NotTo(BeNil())
			})

			It("Should not contain Pages resources", func() {
				Expect(enabled).To(BeFalse())
				Expect(configMap).To(BeNil())
				Expect(service).To(BeNil())
				Expect(deployment).To(BeNil())
				Expect(ingress).To(BeNil())
			})
		})
	})
})

// dumpTemplate() will serialize the template and display the YAML for debugging.
func dumpTemplate(template helm.Template) string { //nolint:golint,unused
	output := new(strings.Builder)

	s := k8sjson.NewYAMLSerializer(k8sjson.DefaultMetaFactory, scheme.Scheme, scheme.Scheme)

	for _, o := range template.Objects() {
		output.WriteString("---\n")
		_ = s.Encode(o, output)
	}

	return output.String()
}

// dumpTemplateToFile() will output the Helm template to a file.
// Note: the file is written to where the test runs NOT from where the
//       tests were run from.
func dumpTemplateToFile(template helm.Template, filename string) error { //nolint:golint,deadcode,unused
	fh, err := os.Create(filename)
	if err != nil {
		return err
	}

	_, _ = fh.WriteString(dumpTemplate(template))

	fh.Close()

	return nil
}

// dumpHelmValues() will output the current values that Helm is using.
func dumpHelmValues(values helm.Values) string { //nolint:golint,unused
	output, _ := json.MarshalIndent(values.AsMap(), "", "    ")
	return string(output)
}

// dumpHelmValuesToFile() will output the current values to a file.
// Note: the file is written to where the test runs NOT from where the
//       tests were run from.
func dumpHelmValuesToFile(values helm.Values, filename string) error { //nolint:golint,deadcode,unused
	fh, err := os.Create(filename)
	if err != nil {
		return err
	}

	_, _ = fh.WriteString(dumpHelmValues(values))

	fh.Close()

	return nil
}
