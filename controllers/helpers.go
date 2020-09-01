package controllers

import (
	gitlabutils "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func getNamespacedName(obj interface{}) types.NamespacedName {
	object := obj.(metav1.Object)
	return types.NamespacedName{Name: object.GetName(), Namespace: object.GetNamespace()}
}

func (r *GitLabReconciler) isObjectFound(object interface{}) bool {
	return gitlabutils.IsObjectFound(r.Client, getNamespacedName(object), object.(runtime.Object))
}

func (r *RunnerReconciler) isObjectFound(object interface{}) bool {
	return gitlabutils.IsObjectFound(r.Client, getNamespacedName(object), object.(runtime.Object))
}

func (r *GLBackupReconciler) isObjectFound(object interface{}) bool {
	return gitlabutils.IsObjectFound(r.Client, getNamespacedName(object), object.(runtime.Object))
}
