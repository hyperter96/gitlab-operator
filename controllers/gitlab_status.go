package controllers

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gitlabctl "gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/gitlab"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
)

const (
	ConditionInitialized = "Initialized"
	ConditionAvailable   = "Available"
	ConditionUpgrading   = "Upgrading"
)

func (r *GitLabReconciler) reconcileGitLabStatus(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, template helm.Template) (ctrl.Result, error) {
	resultRequeue := ctrl.Result{RequeueAfter: 10 * time.Second}
	resultNoRequeue := ctrl.Result{}
	result := resultNoRequeue

	if r.sidekiqAndWebserviceRunning(ctx, adapter, template) {
		adapter.Resource().Status.Phase = "Running"

		if err := r.setStatusCondition(ctx, adapter, ConditionAvailable, true, "GitLab is running and available to accept requests"); err != nil {
			return result, err
		}
	} else {
		adapter.Resource().Status.Phase = "Preparing"
		result = resultRequeue
	}

	// Set the version regardless of whether Sidekiq and Webservice are fully running to
	// ensure we don't trigger the upgrade logic again on the next iteration.
	adapter.Resource().Status.Version = adapter.ChartVersion()

	if err := r.Status().Update(ctx, adapter.Resource()); err != nil {
		return result, err
	}

	time.Sleep(5 * time.Second)

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

func (r *GitLabReconciler) componentRunning(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, deployments []client.Object) bool {
	running := true

	for _, deployment := range deployments {
		// This is only for safeguarding
		if deployment.GetObjectKind().GroupVersionKind().Kind != gitlabctl.DeploymentKind {
			continue
		}

		webservice := &appsv1.Deployment{}
		key := types.NamespacedName{
			Name:      deployment.GetName(),
			Namespace: adapter.Namespace(),
		}

		err := r.Get(ctx, key, webservice)
		if err != nil || !deploymentComplete(webservice, &webservice.Status) {
			running = false
			break
		}
	}

	return running
}

func (r *GitLabReconciler) sidekiqRunning(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, template helm.Template) bool {
	return r.componentRunning(ctx, adapter, gitlabctl.SidekiqDeployments(template))
}

func (r *GitLabReconciler) webserviceRunning(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, template helm.Template) bool {
	return r.componentRunning(ctx, adapter, gitlabctl.WebserviceDeployments(template))
}

func (r *GitLabReconciler) sidekiqAndWebserviceRunning(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, template helm.Template) bool {
	return r.sidekiqRunning(ctx, adapter, template) && r.webserviceRunning(ctx, adapter, template)
}

func (r *GitLabReconciler) sidekiqRunningWithRetry(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, template helm.Template) bool {
	fn := func() error {
		if r.sidekiqRunning(ctx, adapter, template) {
			return nil
		}

		return fmt.Errorf("sidekiq not fully running")
	}

	if err := r.runWithRetry(adapter, fn); err != nil {
		return false
	}

	return true
}

func (r *GitLabReconciler) webserviceRunningWithRetry(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, template helm.Template) bool {
	fn := func() error {
		if r.webserviceRunning(ctx, adapter, template) {
			return nil
		}

		return fmt.Errorf("webservice not fully running")
	}

	if err := r.runWithRetry(adapter, fn); err != nil {
		return false
	}

	return true
}

func (r *GitLabReconciler) runWithRetry(adapter gitlabctl.CustomResourceAdapter, fn func() error) error {
	logger := r.Log.WithValues("gitlab", adapter.Reference(), "namespace", adapter.Namespace())

	time.Sleep(5 * time.Second)

	timeout := 0

	for {
		if timeout > 300 {
			return fmt.Errorf("timeout was longer than 300 seconds")
		}

		err := fn()

		if err != nil {
			logger.V(1).Info(err.Error())

			timeout += 10

			time.Sleep(10 * time.Second)

			continue
		}

		logger.V(1).Info("check passed, proceeding")

		break
	}

	return nil
}

func (r *GitLabReconciler) setStatusCondition(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, reason string, status bool, message string) error {
	statusValue := metav1.ConditionFalse
	if status {
		statusValue = metav1.ConditionTrue
	}

	apimeta.SetStatusCondition(&adapter.Resource().Status.Conditions, metav1.Condition{
		Type:               reason,
		Status:             statusValue,
		Reason:             reason,
		Message:            message,
		ObservedGeneration: adapter.Resource().Generation,
	})

	return r.Status().Update(ctx, adapter.Resource())
}
