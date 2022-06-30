package controllers

import (
	"context"
	"fmt"

	gitlabctl "gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/gitlab"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
)

func (r *GitLabReconciler) reconcileRedis(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, template helm.Template) error {
	if err := r.reconcileRedisConfigMaps(ctx, adapter, template); err != nil {
		return err
	}

	if err := r.reconcileRedisStatefulSet(ctx, adapter, template); err != nil {
		return err
	}

	if err := r.reconcileRedisServices(ctx, adapter, template); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcileRedisConfigMaps(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, template helm.Template) error {
	for _, cm := range gitlabctl.RedisConfigMaps(adapter, template) {
		if err := r.createOrPatch(ctx, cm, adapter); err != nil {
			return err
		}
	}

	return nil
}

func (r *GitLabReconciler) reconcileRedisStatefulSet(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, template helm.Template) error {
	redis := gitlabctl.RedisStatefulSet(template)

	if err := r.annotateSecretsChecksum(ctx, adapter, redis); err != nil {
		return err
	}

	if err := r.createOrPatch(ctx, redis, adapter); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) validateExternalRedisConfiguration(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	defaultRedisSecretName := adapter.Values().GetString("global.redis.password.secret")
	if defaultRedisSecretName == "" {
		defaultRedisSecretName = fmt.Sprintf("%s-%s-secret", adapter.ReleaseName(), gitlabctl.RedisComponentName)
	}

	// If external Redis global password is enabled, ensure it was created.
	if gitlabctl.RedisEnabled(adapter) {
		redisSecretName := adapter.Values().GetString("global.redis.password.secret", defaultRedisSecretName)
		if err := r.ensureSecret(ctx, adapter, redisSecretName); err != nil {
			return err
		}
	}

	// If any of the sub-queues and configured, ensure relevant Secrets are created if enabled.
	for _, subqueue := range gitlabctl.RedisSubqueues() {
		if host := adapter.Values().GetString(fmt.Sprintf("global.redis.%s.host", subqueue)); host != "" {
			// Subqueue is configured. Ensure its password was created.
			if passwordEnabled := adapter.Values().GetBool(fmt.Sprintf("global.redis.%s.password.enabled", subqueue), true); passwordEnabled {
				subqueueSecretName := adapter.Values().GetString(fmt.Sprintf("global.redis.%s.password.secret", subqueue), defaultRedisSecretName)
				if err := r.ensureSecret(ctx, adapter, subqueueSecretName); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (r *GitLabReconciler) reconcileRedisServices(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, template helm.Template) error {
	for _, svc := range gitlabctl.RedisServices(adapter, template) {
		if err := r.createOrPatch(ctx, svc, adapter); err != nil {
			return err
		}
	}

	return nil
}
