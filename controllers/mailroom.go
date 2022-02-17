package controllers

import (
	"context"

	gitlabctl "gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/gitlab"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
)

func (r *GitLabReconciler) reconcileMailroom(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, template helm.Template) error {
	if err := r.reconcileMailroomConfigMap(ctx, adapter, template); err != nil {
		return err
	}

	if err := r.reconcileMailroomDeployment(ctx, adapter, template); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcileMailroomConfigMap(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, template helm.Template) error {
	if _, err := r.createOrPatch(ctx, gitlabctl.MailroomConfigMap(adapter, template), adapter); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcileMailroomDeployment(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, template helm.Template) error {
	deployment := gitlabctl.MailroomDeployment(template)
	if err := r.annotateSecretsChecksum(ctx, adapter, deployment); err != nil {
		return err
	}

	if _, err := r.createOrPatch(ctx, deployment, adapter); err != nil {
		return err
	}

	return nil
}
