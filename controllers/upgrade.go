package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gitlabctl "gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/gitlab"
)

const (
	gitlabLastRestartAnnotationKey = "gitlab.com/last-restart"
	timeFormat                     = "20060102150405"
	envVarNameBypassSchemaVersion  = "BYPASS_SCHEMA_VERSION" //nolint:gosec // for some reason this is suspected as an exposed credential
	initContainerNameDependencies  = "dependencies"
)

func (r *GitLabReconciler) getDeployment(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, deploymentName string) (*appsv1.Deployment, error) {
	deployment := &appsv1.Deployment{}
	lookupKey := types.NamespacedName{Namespace: adapter.Namespace(), Name: deploymentName}

	if err := r.Get(ctx, lookupKey, deployment); err != nil {
		return deployment, fmt.Errorf("unable to get Deployment: %s", err.Error())
	}

	return deployment, nil
}

func (r *GitLabReconciler) unpauseDeployments(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, deployments []client.Object) error {
	for i := range deployments {
		deployment, err := r.getDeployment(ctx, adapter, deployments[i].GetName())
		if err != nil {
			return err
		}

		deployment.Spec.Paused = false

		// If unpausing during an upgrade, then set BYPASS_SCHEMA_VERSION.
		if adapter.IsUpgrade() {
			for j := range deployment.Spec.Template.Spec.InitContainers {
				if deployment.Spec.Template.Spec.InitContainers[j].Name == initContainerNameDependencies {
					deployment.Spec.Template.Spec.InitContainers[j].Env = append(
						deployment.Spec.Template.Spec.InitContainers[j].Env,
						corev1.EnvVar{Name: envVarNameBypassSchemaVersion, Value: "true"})
				}
			}
		}

		err = r.Update(ctx, deployment)
		if err != nil {
			return fmt.Errorf("unable to update deployment %s: %s", deployment.Name, err.Error())
		}
	}

	return nil
}

func (r *GitLabReconciler) unpauseWebserviceDeployments(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	return r.unpauseDeployments(ctx, adapter, gitlabctl.WebserviceDeployments(adapter))
}

func (r *GitLabReconciler) unpauseSidekiqDeployments(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	return r.unpauseDeployments(ctx, adapter, gitlabctl.SidekiqDeployments(adapter))
}

func (r *GitLabReconciler) rollingUpdateDeployments(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, deployments []client.Object) error {
	for i := range deployments {
		deployment, err := r.getDeployment(ctx, adapter, deployments[i].GetName())
		if err != nil {
			return err
		}

		deployment.Spec.Template.ObjectMeta.Annotations[gitlabLastRestartAnnotationKey] = time.Now().Format(timeFormat)
		removeInitContainerEnvVar(deployment, initContainerNameDependencies, envVarNameBypassSchemaVersion)

		if err := r.Update(ctx, deployment); err != nil {
			return err
		}
	}

	return nil
}

func (r *GitLabReconciler) rollingUpdateWebserviceDeployments(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	return r.rollingUpdateDeployments(ctx, adapter, gitlabctl.WebserviceDeployments(adapter))
}

func (r *GitLabReconciler) rollingUpdateSidekiqDeployments(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	return r.rollingUpdateDeployments(ctx, adapter, gitlabctl.SidekiqDeployments(adapter))
}

func (r *GitLabReconciler) reconcileWebserviceAndSidekiqIfEnabled(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, pause bool, log logr.Logger) error {
	if gitlabctl.WebserviceEnabled(adapter) {
		log.Info("reconciling Webservice Deployments", "pause", pause)

		if err := r.reconcileWebserviceDeployments(ctx, adapter, pause); err != nil {
			return err
		}
	}

	if gitlabctl.SidekiqEnabled(adapter) {
		log.Info("reconciling Sidekiq Deployments", "pause", pause)

		if err := r.reconcileSidekiqDeployments(ctx, adapter, pause); err != nil {
			return err
		}
	}

	return nil
}

func (r *GitLabReconciler) unpauseWebserviceAndSidekiqIfEnabled(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, log logr.Logger) error {
	if gitlabctl.WebserviceEnabled(adapter) {
		log.Info("ensuring Webservice Deployments are unpaused")

		if err := r.unpauseWebserviceDeployments(ctx, adapter); err != nil {
			return err
		}
	}

	if gitlabctl.SidekiqEnabled(adapter) {
		log.Info("ensuring Sidekiq Deployments are unpaused")

		if err := r.unpauseSidekiqDeployments(ctx, adapter); err != nil {
			return err
		}
	}

	return nil
}

func (r *GitLabReconciler) webserviceAndSidekiqRunningIfEnabled(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, log logr.Logger) error {
	if gitlabctl.WebserviceEnabled(adapter) {
		log.Info("ensuring Webservice Deployments are running")

		if !r.webserviceRunningWithRetry(ctx, adapter) {
			return fmt.Errorf("Webservice has not started fully")
		}
	}

	if gitlabctl.SidekiqEnabled(adapter) {
		log.Info("ensuring Sidekiq Deployments are running")

		if !r.sidekiqRunningWithRetry(ctx, adapter) {
			return fmt.Errorf("Sidekiq has not started fully")
		}
	}

	return nil
}

func (r *GitLabReconciler) rollingUpdateWebserviceAndSidekiqIfEnabled(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, log logr.Logger) error {
	if gitlabctl.WebserviceEnabled(adapter) {
		log.Info("ensuring Webservice Deployments are running")

		if err := r.rollingUpdateWebserviceDeployments(ctx, adapter); err != nil {
			return err
		}
	}

	if gitlabctl.SidekiqEnabled(adapter) {
		log.Info("ensuring Sidekiq Deployments are running")

		if err := r.rollingUpdateSidekiqDeployments(ctx, adapter); err != nil {
			return err
		}
	}

	return nil
}

func removeInitContainerEnvVar(deployment *appsv1.Deployment, initContainerName, envVarName string) {
	for i := range deployment.Spec.Template.Spec.InitContainers {
		if deployment.Spec.Template.Spec.InitContainers[i].Name == initContainerName {
			for j := range deployment.Spec.Template.Spec.InitContainers[i].Env {
				if deployment.Spec.Template.Spec.InitContainers[i].Env[j].Name == envVarName {
					deployment.Spec.Template.Spec.InitContainers[i].Env = removeEnvVar(deployment.Spec.Template.Spec.InitContainers[i].Env, j)
				}
			}
		}
	}
}

func removeEnvVar(vars []corev1.EnvVar, i int) []corev1.EnvVar {
	// copy the last element to the spot we want to remove
	vars[i] = vars[len(vars)-1]

	// remove the last ("copied") element from the slice
	return vars[:len(vars)-1]
}
