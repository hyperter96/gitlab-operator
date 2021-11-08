package controllers

import (
	"context"

	gitlabctl "gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/gitlab"
)

func (r *GitLabReconciler) reconcileMailroom(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	if err := r.reconcileMailroomConfigMap(ctx, adapter); err != nil {
		return err
	}

	if err := r.reconcileMailroomDeployment(ctx, adapter); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcileMailroomConfigMap(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	if _, err := r.createOrPatch(ctx, gitlabctl.MailroomConfigMap(adapter), adapter); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcileMailroomDeployment(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	deployment := gitlabctl.MailroomDeployment(adapter)
	if err := r.annotateSecretsChecksum(ctx, adapter, &deployment.Spec.Template); err != nil {
		return err
	}

	if _, err := r.createOrPatch(ctx, deployment, adapter); err != nil {
		return err
	}

	return nil
}
