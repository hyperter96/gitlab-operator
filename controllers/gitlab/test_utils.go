package gitlab

import (
	gitlabv1beta1 "gitlab.com/gitlab-org/cloud-native/gitlab-operator/api/v1beta1"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createMockAdapter(namespace string, version string, values helm.Values) CustomResourceAdapter {
	mockGitLab := &gitlabv1beta1.GitLab{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: namespace,
		},
		Spec: gitlabv1beta1.GitLabSpec{
			Chart: gitlabv1beta1.GitLabChartSpec{
				Version: version,
				Values: gitlabv1beta1.ChartValues{
					Object: values.AsMap(),
				},
			},
		},
	}

	adapter := NewCustomResourceAdapter(mockGitLab)

	return adapter
}

