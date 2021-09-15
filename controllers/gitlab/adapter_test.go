package gitlab

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	gitlabv1beta1 "gitlab.com/gitlab-org/cloud-native/gitlab-operator/api/v1beta1"
)

var _ = Describe("CustomResourceAdapter", func() {

	if namespace == "" {
		namespace = "default" //nolint:golint,goconst
	}

	mockGitLab := &gitlabv1beta1.GitLab{
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

	It("retrieve the attributes from GitLab CR", func() {

		adapter := NewCustomResourceAdapter(mockGitLab)

		Expect(adapter.Reference()).To(Equal(fmt.Sprintf("test.%s", namespace)))
		Expect(adapter.Namespace()).To(Equal(namespace))
		Expect(adapter.ReleaseName()).To(Equal("test"))
		Expect(adapter.ChartVersion()).To(Equal(chartVersions[0]))
	})

	It("should change the hash when values change", func() {

		adapter := NewCustomResourceAdapter(mockGitLab)

		gitlabCopy := mockGitLab.DeepCopy()

		gitlabCopy.Spec.Chart.Values.Object = map[string]interface{}{
			"foo": "FOO",
			"bar": map[string]interface{}{
				"baz": "BAZ",
			},
		}

		beforeHash := adapter.Hash()

		adapter = NewCustomResourceAdapter(gitlabCopy)

		afterHash := adapter.Hash()

		Expect(beforeHash).NotTo(Equal(afterHash))
	})

})
