package controllers

import (
	"context"

	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
	gitlabutils "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
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

func (r *GitLabReconciler) isEndpointReady(ctx context.Context, service string, cr *gitlabv1beta1.GitLab) bool {
	var addresses []corev1.EndpointAddress

	ep := &corev1.Endpoints{}
	err := r.Get(ctx, types.NamespacedName{Name: service, Namespace: cr.Namespace}, ep)
	if err != nil && errors.IsNotFound(err) {
		return false
	}

	for _, subset := range ep.Subsets {
		addresses = append(addresses, subset.Addresses...)
	}

	return len(addresses) > 0
}

func (r *GitLabReconciler) ifCoreServicesReady(ctx context.Context, cr *gitlabv1beta1.GitLab) bool {
	return r.isEndpointReady(ctx, cr.Name+"-postgresql", cr) &&
		r.isEndpointReady(ctx, cr.Name+"-gitaly", cr) &&
		r.isEndpointReady(ctx, cr.Name+"-redis-master", cr)
}

func getLabelSet(cr *gitlabv1beta1.GitLab) labels.Set {
	webLabels := gitlabutils.Label(cr.Name, "webservice", gitlabutils.GitlabType)

	unwantedKeys := []string{"app.kubernetes.io/component", "app.kubernetes.io/instance"}
	for _, key := range unwantedKeys {
		delete(webLabels, key)
	}

	return labels.Set(webLabels)
}
