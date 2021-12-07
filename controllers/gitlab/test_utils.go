package gitlab

import (
	"os"

	gitlabv1beta1 "gitlab.com/gitlab-org/cloud-native/gitlab-operator/api/v1beta1"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	releaseName = "test"
	namespace   = os.Getenv("HELM_NAMESPACE")
)

func GetChartVersion() string {
	version, found := os.LookupEnv("CHART_VERSION")
	if !found {
		version = AvailableChartVersions()[0]
	}

	return version
}

func CreateMockGitLab(releaseName, namespace string, values helm.Values) *gitlabv1beta1.GitLab {
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
				Version: GetChartVersion(),
				Values: gitlabv1beta1.ChartValues{
					Object: values.AsMap(),
				},
			},
		},
	}
}

func CreateMockAdapter(mockGitLab *gitlabv1beta1.GitLab) CustomResourceAdapter {
	return NewCustomResourceAdapter(mockGitLab)
}
