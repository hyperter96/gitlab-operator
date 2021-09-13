package controllers

import (
	"context"
	"reflect"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gitlabv1beta1 "gitlab.com/gitlab-org/cloud-native/gitlab-operator/api/v1beta1"
	gitlabctl "gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/gitlab"
)

func (r *GitLabReconciler) reconcileGitlabStatus(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) (ctrl.Result, error) {
	waitInterval := 5 * time.Second
	resultRequeue := ctrl.Result{RequeueAfter: waitInterval}
	resultNoRequeue := ctrl.Result{}
	lookupKey := types.NamespacedName{Namespace: adapter.Namespace(), Name: adapter.ReleaseName()}

	// get current Gitlab resource
	gitlab := &gitlabv1beta1.GitLab{}
	if err := r.Get(ctx, lookupKey, gitlab); err != nil {
		return resultRequeue, err
	}

	result := resultNoRequeue

	if r.isWebserviceRunning(ctx, adapter) {
		gitlab.Status.Phase = "Running"
		gitlab.Status.Stage = ""
	} else {
		gitlab.Status.Phase = "Initializing"
		gitlab.Status.Stage = "Gitlab is initializing"

		r.Log.V(1).Info("webservice not fully running, will wait and retry", "interval", waitInterval)

		result = resultRequeue
	}

	// Check if the status of the gitlab resource has changed
	if !reflect.DeepEqual(adapter.Resource().Status, gitlab.Status) {
		// Update status if the status has changed
		if err := r.setGitlabStatus(ctx, gitlab); err != nil {
			return resultRequeue, err
		}
	}

	return result, nil
}

// Same check as used in the deployment utils in upstream Kubernetes
// https://github.com/kubernetes/kubernetes/blob/master/pkg/controller/deployment/util/deployment_util.go#L722
func deploymentComplete(deployment *appsv1.Deployment, newStatus *appsv1.DeploymentStatus) bool {
	return newStatus.UpdatedReplicas == *(deployment.Spec.Replicas) &&
		newStatus.Replicas == *(deployment.Spec.Replicas) &&
		newStatus.AvailableReplicas == *(deployment.Spec.Replicas) &&
		newStatus.ObservedGeneration >= deployment.Generation
}

func (r *GitLabReconciler) isWebserviceRunning(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) bool {
	running := false

	// confirm that all Webservice deployments are running
	for _, deployment := range gitlabctl.WebserviceDeployments(adapter) {
		webservice := &appsv1.Deployment{}
		key := types.NamespacedName{
			Name:      deployment.Name,
			Namespace: adapter.Namespace(),
		}

		err := r.Get(ctx, key, webservice)
		if err == nil && deploymentComplete(webservice, &webservice.Status) {
			running = true
		}
	}

	return running
}

// setGitlabStatus sets status of custom resource.
func (r *GitLabReconciler) setGitlabStatus(ctx context.Context, object client.Object) error {
	return r.Status().Update(ctx, object)
}
