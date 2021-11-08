package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"

	gitlabctl "gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/gitlab"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/internal"
)

func (r *GitLabReconciler) reconcileMinioInstance(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	cm := internal.MinioScriptConfigMap(adapter)
	if _, err := r.createOrPatch(ctx, cm, adapter); err != nil {
		return err
	}

	secret := internal.MinioSecret(adapter)
	if _, err := r.createOrPatch(ctx, secret, adapter); err != nil && errors.IsAlreadyExists(err) {
		return err
	}

	appConfigSecret, err := internal.AppConfigConnectionSecret(adapter, *secret)
	if err != nil {
		return err
	}

	if _, err := r.createOrPatch(ctx, appConfigSecret, adapter); err != nil && errors.IsAlreadyExists(err) {
		return err
	}

	registryConnectionSecret, err := internal.RegistryConnectionSecret(adapter, *secret)
	if err != nil {
		return err
	}

	if _, err := r.createOrPatch(ctx, registryConnectionSecret, adapter); err != nil && errors.IsAlreadyExists(err) {
		return err
	}

	toolboxConnectionSecret := internal.ToolboxConnectionSecret(adapter, *secret)
	if _, err := r.createOrPatch(ctx, toolboxConnectionSecret, adapter); err != nil && errors.IsAlreadyExists(err) {
		return err
	}

	svc := internal.MinioService(adapter)
	if _, err := r.createOrPatch(ctx, svc, adapter); err != nil {
		return err
	}

	minio := internal.MinioStatefulSet(adapter)
	if err := r.annotateSecretsChecksum(ctx, adapter, &minio.Spec.Template); err != nil {
		return err
	}

	_, err = r.createOrPatch(ctx, minio, adapter)
	if err != nil {
		return err
	}

	buckets := internal.BucketCreationJob(adapter)
	if _, err := r.createOrPatch(ctx, buckets, adapter); err != nil {
		return err
	}

	ingress := internal.MinioIngress(adapter)
	if _, err = r.createOrPatch(ctx, ingress, adapter); err != nil {
		return err
	}

	return nil
}
