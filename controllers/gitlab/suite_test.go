package gitlab

import (
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/helpers"
	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/settings"
)

func GitLabMock() *gitlabv1beta1.GitLab {
	releaseName := "test"
	chartVersion := helpers.AvailableChartVersions()[0]
	namespace := os.Getenv("HELM_NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}

	return &gitlabv1beta1.GitLab{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps.gitlab.com/v1beta1",
			Kind:       "GitLab",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      releaseName,
			Namespace: namespace,
			Labels: map[string]string{
				"chart": fmt.Sprintf("gitlab-%s", chartVersion),
			},
		},
		Spec: gitlabv1beta1.GitLabSpec{
			AutoScaling: &gitlabv1beta1.AutoScalingSpec{},
			Database: &gitlabv1beta1.DatabaseSpec{
				Volume: gitlabv1beta1.VolumeSpec{
					Capacity: "50Gi",
				},
			},
			Redis: &gitlabv1beta1.RedisSpec{},
			Volume: gitlabv1beta1.VolumeSpec{
				Capacity: "50Gi",
			},
		},
	}
}

func TestGitLab(t *testing.T) {
	settings.Load()
	RegisterFailHandler(Fail)
	RunSpecs(t, "GitLab Suite")
}
