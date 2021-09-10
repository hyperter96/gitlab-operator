package gitlab

import (
	"os"
	"strings"

	gitlabv1beta1 "gitlab.com/gitlab-org/cloud-native/gitlab-operator/api/v1beta1"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"

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
func dumpTemplate(template helm.Template) string {
	output := new(strings.Builder)

	s := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme.Scheme, scheme.Scheme)
	for _, o := range(template.Objects()) {
		output.WriteString("---\n")
		s.Encode(o, output)
	}

	return output.String()
}

// dumpTemplateToFile() will output the Helm template to a file
// Note: the file is written to where the test runs NOT from where the
//       tests were run from
func dumpTemplateToFile(template helm.Template, filename string) error {
	fh, err := os.Create(filename)
	if err != nil {
		return err
	}
	fh.WriteString(dumpTemplate(template))
	fh.Close()
	return nil
}
