package gitlab_test

import (
	"fmt"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/onsi/gomega/types"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func SatisfyReplacement(expected interface{}, fieldsToIgnore ...*FieldsToIgnore) types.GomegaMatcher {
	return &internalReplacementMatcher{
		expected:       expected,
		fieldsToIgnore: fieldsToIgnore,
	}
}

func IgnoreFields(kind interface{}, fields ...string) *FieldsToIgnore {
	return &FieldsToIgnore{
		kind:   kind,
		fields: fields,
	}
}

type FieldsToIgnore struct {
	kind   interface{}
	fields []string
}

type internalReplacementMatcher struct {
	expected       interface{}
	differences    string
	fieldsToIgnore []*FieldsToIgnore
}

func (m *internalReplacementMatcher) Match(actual interface{}) (success bool, err error) {
	cmpOptions := []cmp.Option{
		cmpopts.IgnoreFields(metav1.ObjectMeta{}, negligibleObjectMetaFields...),
		cmpopts.IgnoreFields(appsv1.Deployment{}, commonFieldsToIgnore...),
		cmpopts.IgnoreFields(appsv1.DeploymentSpec{}, negligibleDeploymentSpecFields...),
	}

	for _, entry := range m.fieldsToIgnore {
		cmpOptions = append(cmpOptions, cmpopts.IgnoreFields(entry.kind, entry.fields...))
	}

	m.differences = cmp.Diff(m.expected, actual, cmpOptions...)
	return cmp.Equal(m.expected, actual, cmpOptions...), nil
}

func (m *internalReplacementMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected a satisfactory replacement but found significant differences.\n\n"+
		"--- %T (expected)\n"+
		"+++ %T (actual)\n\n"+
		"%s", m.expected, actual, m.differences)
}

func (m *internalReplacementMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%#v\nnot to satisfy\n\t%#v", actual, m.expected)
}

var (
	negligibleObjectMetaFields = []string{
		"GenerateName",
		"SelfLink",
		"UID",
		"ResourceVersion",
		"Generation",
		"CreationTimestamp",
		"DeletionTimestamp",
		"DeletionGracePeriodSeconds",
		"OwnerReferences",
		"Finalizers",
		"ClusterName",
		"ManagedFields",
	}
	negligibleDeploymentSpecFields = []string{
		"MinReadySeconds",
		"RevisionHistoryLimit",
		"ProgressDeadlineSeconds",
	}
	commonFieldsToIgnore = []string{
		"TypeMeta",
		"Status",
	}
)
