package gitlab

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Status represents the status of the underlying GitLab resource and
// provides a convenient mechanism to update the status.
type Status interface {
	// SetCondition updates the provided condition with the details of
	// the underlying GitLab resource, including the observed generation,
	// and adds it to the resource conditions.
	SetCondition(condition metav1.Condition)

	// RecordVersion sets the status version to the specified (desired) version.
	RecordVersion()
}
