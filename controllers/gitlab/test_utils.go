package gitlab

import (
	"os"

	gitlabv1beta1 "gitlab.com/gitlab-org/cloud-native/gitlab-operator/api/v1beta1"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/resource"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	releaseName = "test"
	namespace   = os.Getenv("HELM_NAMESPACE")
)

func CreateMockGitLab(releaseName, namespace string, values resource.Values) *gitlabv1beta1.GitLab {
	return &gitlabv1beta1.GitLab{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps.gitlab.com/v1beta1",
			Kind:       "GitLab",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      releaseName,
			Namespace: namespace,
		},
		Spec: gitlabv1beta1.GitLabSpec{
			Chart: gitlabv1beta1.GitLabChartSpec{
				Version: helm.GetChartVersion(),
				Values: gitlabv1beta1.ChartValues{
					Object: values,
				},
			},
		},
	}
}

func CreateMockAdapter(mockGitLab *gitlabv1beta1.GitLab) CustomResourceAdapter {
	return NewCustomResourceAdapter(mockGitLab)
}
