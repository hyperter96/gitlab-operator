package gitlab

import (
	"os"
	"syscall"

	gitlabv1beta1 "gitlab.com/gitlab-org/cloud-native/gitlab-operator/api/v1beta1"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"

	//corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes/scheme"
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

// dumpTemplate() will serialize the template and display the YAML for debugging
func dumpTemplate(adapter CustomResourceAdapter) {
	stdout := os.NewFile(uintptr(syscall.Stdout), "/dev/stdout")
	s := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme.Scheme, scheme.Scheme)
	template, _ := GetTemplate(adapter)
	for _, o := range(template.Objects()) {
		stdout.WriteString("---\n")
		s.Encode(o, stdout)
	}
}
