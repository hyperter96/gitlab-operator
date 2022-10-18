package v1beta1

import (
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/gitlab/status"
)

/* GitLabStatus */

func (w *Adapter) SetCondition(condition metav1.Condition) {
	condition.ObservedGeneration = w.source.Generation

	/* Deduce phase from condition type */
	if condition.Type == status.ConditionAvailable.Name() {
		if condition.Status == metav1.ConditionTrue {
			w.source.Status.Phase = status.PhaseRunning
		}
	} else {
		w.source.Status.Phase = status.PhasePreparing
	}

	meta.SetStatusCondition(&w.source.Status.Conditions, condition)
}

func (w *Adapter) RecordVersion() {
	w.source.Status.Version = w.DesiredVersion()
}
