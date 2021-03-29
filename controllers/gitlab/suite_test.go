package gitlab

import (
	"context"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/helpers"
	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/settings"
	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/helm"
)

var (
	ctx          = context.Background()
	chartVersion = helpers.AvailableChartVersions()[0]
	emptyValues  = helm.EmptyValues()
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
					Object: emptyValues.AsMap(),
				},
			}},
	}
}

// CfgMapFromList returns a ConfigMap by name from a list of ConfigMaps.
func CfgMapFromList(name string, cfgMaps []*corev1.ConfigMap) *corev1.ConfigMap {
	for _, cm := range cfgMaps {
		if cm.Name == name {
			return cm
		}
	}

	return nil
}

// SvcFromList returns a Service by name from a list of Services.
func SvcFromList(name string, services []*corev1.Service) *corev1.Service {
	for _, s := range services {
		if s.Name == name {
			return s
		}
	}

	return nil
}

func TestGitLab(t *testing.T) {
	settings.Load()
	RegisterFailHandler(Fail)
	RunSpecs(t, "GitLab Suite")
}
