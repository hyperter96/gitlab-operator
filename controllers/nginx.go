package controllers

import (
	"context"

	gitlabctl "gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/gitlab"
)

func (r *GitLabReconciler) reconcileNGINX(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	for _, cm := range gitlabctl.NGINXConfigMaps(adapter) {
		if _, err := r.createOrPatch(ctx, cm, adapter); err != nil {
			return err
		}
	}

	for _, svc := range gitlabctl.NGINXServices(adapter) {
		if _, err := r.createOrPatch(ctx, svc, adapter); err != nil {
			return err
		}
	}

	for _, dep := range gitlabctl.NGINXDeployments(adapter) {
		if _, err := r.createOrPatch(ctx, dep, adapter); err != nil {
			return err
		}
	}

	return nil
}
