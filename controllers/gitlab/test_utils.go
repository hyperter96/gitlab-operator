package gitlab

import (
	"encoding/json"
	"os"
	"strings"

	gitlabv1beta1 "gitlab.com/gitlab-org/cloud-native/gitlab-operator/api/v1beta1"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sjson "k8s.io/apimachinery/pkg/runtime/serializer/json"
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

// dumpTemplate() will serialize the template and display the YAML for debugging.
func dumpTemplate(template helm.Template) string { //nolint:golint,unused
	output := new(strings.Builder)

	s := k8sjson.NewYAMLSerializer(k8sjson.DefaultMetaFactory, scheme.Scheme, scheme.Scheme)

	for _, o := range template.Objects() {
		output.WriteString("---\n")
		_ = s.Encode(o, output)
	}

	return output.String()
}

// dumpTemplateToFile() will output the Helm template to a file.
// Note: the file is written to where the test runs NOT from where the
//       tests were run from.
func dumpTemplateToFile(template helm.Template, filename string) error { //nolint:golint,deadcode
	fh, err := os.Create(filename)
	if err != nil {
		return err
	}

	_, _ = fh.WriteString(dumpTemplate(template))

	fh.Close()

	return nil
}

// dumpHelmValues() will output the current values that Helm is using.
func dumpHelmValues(values helm.Values) string { //nolint:golint,unused
	output, _ := json.MarshalIndent(values.AsMap(), "", "    ")
	return string(output)
}

// dumpHelmValuesToFile() will output the current values to a file.
// Note: the file is written to where the test runs NOT from where the
//       tests were run from.
func dumpHelmValuesToFile(values helm.Values, filename string) error { //nolint:golint,deadcode
	fh, err := os.Create(filename)
	if err != nil {
		return err
	}

	_, _ = fh.WriteString(dumpHelmValues(values))

	fh.Close()

	return nil
}
