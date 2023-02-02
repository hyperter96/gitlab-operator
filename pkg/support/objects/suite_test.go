package objects

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func newObject(apiVersion, kind string) *unstructured.Unstructured {
	r := &unstructured.Unstructured{}

	r.SetAPIVersion(apiVersion)
	r.SetKind(kind)

	return r
}

func setQualifiedName(object *unstructured.Unstructured, name, namespace string) *unstructured.Unstructured {
	object.SetName(name)
	object.SetNamespace(namespace)

	return object
}

func setLabels(object *unstructured.Unstructured, labels map[string]string) *unstructured.Unstructured {
	object.SetLabels(labels)

	return object
}

func setAnnotations(object *unstructured.Unstructured, annotations map[string]string) *unstructured.Unstructured {
	object.SetAnnotations(annotations)

	return object
}

func TestGitlabOperator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GitLab Operator Framework: Objects Support")
}
