package gitlab_test

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/gitlab"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("CustomResourceAdapter", func() {

	namespace := os.Getenv("HELM_NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}

	chartVersions := gitlab.AvailableChartVersions()

	mockGitLab := &gitlabv1beta1.GitLab{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: namespace,
			Labels: map[string]string{
				"chart": fmt.Sprintf("gitlab-%s", chartVersions[0]),
			},
		},
	}

	It("retrieve the attributes from GitLab CR", func() {

		adapter := gitlab.NewCustomResourceAdapter(mockGitLab)

		Expect(adapter.Reference()).To(Equal(fmt.Sprintf("test.%s", namespace)))
		Expect(adapter.Namespace()).To(Equal(namespace))
		Expect(adapter.ReleaseName()).To(Equal("test"))
		Expect(adapter.ChartVersion()).To(Equal(chartVersions[0]))
	})

})
