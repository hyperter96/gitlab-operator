package gitlab

import (
	"context"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	gitlabv1beta1 "gitlab.com/gitlab-org/cloud-native/gitlab-operator/api/v1beta1"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/settings"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/gitlab"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/gitlab/adapter"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/support"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/support/charts"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/support/charts/populate"
)

const (
	testNamespace = "default"
)

var (
	releaseName = "test"
	namespace   = os.Getenv("HELM_NAMESPACE")
)

func CreateMockGitLab(releaseName, namespace string, values support.Values) *gitlabv1beta1.GitLab {
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

func CreateMockAdapter(mockGitLab *gitlabv1beta1.GitLab) gitlab.Adapter {
	adapter, _ := adapter.NewV1Beta1(context.TODO(), mockGitLab)

	return adapter
}

func TestGitLab(t *testing.T) {
	settings.Load()
	_ = charts.PopulateGlobalCatalog(
		populate.WithSearchPath(settings.HelmChartsDirectory))

	RegisterFailHandler(Fail)
	RunSpecs(t, "GitLab Suite")
}
