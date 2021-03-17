package gitlab

import (
	"context"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/helpers"
	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/settings"
	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/helm"
)

var (
	ctx          = context.Background()
	chartVersion = helpers.AvailableChartVersions()[0]
	chartValues  = helm.EmptyValues()
	namespace    = os.Getenv("HELM_NAMESPACE")
	releaseName  = "test"
)

func GitLabMock() *gitlabv1beta1.GitLab {
	if namespace == "" {
		namespace = "default"
	}

	// Set chart values

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
				Version: chartVersion,
				Values: gitlabv1beta1.ChartValues{
					Object: chartValues.AsMap(),
				},
			}},
	}
}

func TestGitLab(t *testing.T) {
	settings.Load()
	RegisterFailHandler(Fail)
	RunSpecs(t, "GitLab Suite")
}
