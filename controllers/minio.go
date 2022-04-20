package controllers

import (
	"context"

	gitlabctl "gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/gitlab"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
)

func (r *GitLabReconciler) reconcileMinioInstance(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, template helm.Template) error {
	cm := gitlabctl.MinioConfigMap(adapter, template)
	if _, err := r.createOrPatch(ctx, cm, adapter); err != nil {
		return err
	}

	buckets := gitlabctl.MinioJob(adapter, template)
	if _, err := r.createOrPatch(ctx, buckets, adapter); err != nil {
		return err
	}

	svc := gitlabctl.MinioService(adapter, template)
	if _, err := r.createOrPatch(ctx, svc, adapter); err != nil {
		return err
	}

	pvc := gitlabctl.MinioPersistentVolumeClaim(adapter, template)
	if _, err := r.createOrPatch(ctx, pvc, adapter); err != nil {
		return err
	}

	minio := gitlabctl.MinioDeployment(adapter, template)
	if err := r.annotateSecretsChecksum(ctx, adapter, minio); err != nil {
		return err
	}

	if _, err := r.createOrPatch(ctx, minio, adapter); err != nil {
		return err
	}

	ingress := gitlabctl.MinioIngress(adapter, template)
	if err := r.reconcileIngress(ctx, ingress, adapter); err != nil {
		return err
	}

	return nil
}
