package controllers

import (
	"context"

	gitlabctl "gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/gitlab"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/internal"
)

func (r *GitLabReconciler) reconcileWebserviceExceptDeployments(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	if err := r.reconcileWebserviceConfigMaps(ctx, adapter); err != nil {
		return err
	}

	if err := r.reconcileWebserviceServices(ctx, adapter); err != nil {
		return err
	}

	if err := r.reconcileWebserviceIngresses(ctx, adapter); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcileWebserviceConfigMaps(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	for _, cm := range gitlabctl.WebserviceConfigMaps(adapter) {
		if _, err := r.createOrPatch(ctx, cm, adapter); err != nil {
			return err
		}
	}

	return nil
}

func (r *GitLabReconciler) reconcileWebserviceServices(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	for _, svc := range gitlabctl.WebserviceServices(adapter) {
		if _, err := r.createOrPatch(ctx, svc, adapter); err != nil {
			return err
		}
	}

	return nil
}

func (r *GitLabReconciler) reconcileWebserviceDeployments(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, pause bool) error {
	logger := r.Log.WithValues("gitlab", adapter.Reference(), "namespace", adapter.Namespace())

	webservices := gitlabctl.WebserviceDeployments(adapter)

	if internal.IsOpenshift() && len(webservices) > 1 {
		logger.V(2).Info("Multiple Webservice Ingresses detected, which is not supported on OpenShift when using NGINX Ingress Operator. See https://gitlab.com/gitlab-org/cloud-native/gitlab-operator/-/issues/160")
	}

	for _, webservice := range webservices {
		if err := r.setDeploymentReplica(ctx, webservice); err != nil {
			return err
		}

		if err := r.annotateSecretsChecksum(ctx, adapter, &webservice.Spec.Template); err != nil {
			return err
		}

		if pause {
			webservice.Spec.Paused = true
		} else {
			webservice.Spec.Paused = false
		}

		if _, err := r.createOrPatch(ctx, webservice, adapter); err != nil {
			return err
		}
	}

	return nil
}

func (r *GitLabReconciler) reconcileWebserviceIngresses(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	for _, ingress := range gitlabctl.WebserviceIngresses(adapter) {
		if err := r.reconcileIngress(ctx, ingress, adapter); err != nil {
			return err
		}
	}

	return nil
}
