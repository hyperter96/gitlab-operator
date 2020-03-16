package gitlab

import (
	"context"

	gitlabv1beta1 "github.com/OchiengEd/gitlab-operator/pkg/apis/gitlab/v1beta1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *ReconcileGitlab) reconcileConfigMaps(cr *gitlabv1beta1.Gitlab, s security) error {
	gitlab := getGitlabConfig(cr)

	if IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: gitlab.Name}, gitlab) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, gitlab, r.scheme); err != nil {
		return err
	}

	if err := r.client.Create(context.TODO(), gitlab); err != nil {
		return err
	}

	redis := getRedisConfig(cr, s)

	if IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: redis.Name}, redis) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, redis, r.scheme); err != nil {
		return err
	}

	if err := r.client.Create(context.TODO(), redis); err != nil {
		return err
	}

	return nil
}

func (r *ReconcileGitlab) reconcileSecrets(cr *gitlabv1beta1.Gitlab, s security) error {

	core := getGilabSecret(cr, s)

	if IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: core.Name}, core) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, core, r.scheme); err != nil {
		return err
	}

	if err := r.client.Create(context.TODO(), core); err != nil {
		return err
	}

	return nil
}

func (r *ReconcileGitlab) reconcileServices(cr *gitlabv1beta1.Gitlab) error {
	postgres := getPostgresService(cr)

	if IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: postgres.Name}, postgres) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, postgres, r.scheme); err != nil {
		return err
	}

	if err := r.client.Create(context.TODO(), postgres); err != nil {
		return err
	}

	redis := getRedisService(cr)

	if IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: redis.Name}, redis) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, redis, r.scheme); err != nil {
		return err
	}

	if err := r.client.Create(context.TODO(), redis); err != nil {
		return err
	}

	gitlab := getGitlabService(cr)

	if IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: gitlab.Name}, gitlab) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, gitlab, r.scheme); err != nil {
		return err
	}

	if err := r.client.Create(context.TODO(), gitlab); err != nil {
		return err
	}

	exporter := getMetricsService(cr)

	if IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: exporter.Name}, exporter) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, exporter, r.scheme); err != nil {
		return err
	}

	if err := r.client.Create(context.TODO(), exporter); err != nil {
		return err
	}

	return nil
}

func (r *ReconcileGitlab) reconcilePersistentVolumeClaims(cr *gitlabv1beta1.Gitlab) error {
	if cr.Spec.Registry.Enabled && cr.Spec.Volumes.Registry.Capacity != "" {
		registryVolume := getRegistryVolumeClaim(cr)

		if IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: registryVolume.Name}, registryVolume) {
			return nil
		}

		if err := controllerutil.SetControllerReference(cr, registryVolume, r.scheme); err != nil {
			return err
		}

		if err := r.client.Create(context.TODO(), registryVolume); err != nil {
			return err
		}
	}

	if cr.Spec.Volumes.Data.Capacity != "" {
		dataVolume := getGitlabDataVolumeClaim(cr)

		if IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: dataVolume.Name}, dataVolume) {
			return nil
		}

		if err := controllerutil.SetControllerReference(cr, dataVolume, r.scheme); err != nil {
			return err
		}

		if err := r.client.Create(context.TODO(), dataVolume); err != nil {
			return err
		}
	}

	if cr.Spec.Volumes.Configuration.Capacity != "" {
		configVolume := getGitlabConfigVolumeClaim(cr)

		if IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: configVolume.Name}, configVolume) {
			return nil
		}

		if err := controllerutil.SetControllerReference(cr, configVolume, r.scheme); err != nil {
			return err
		}

		if err := r.client.Create(context.TODO(), configVolume); err != nil {
			return err
		}
	}

	return nil
}

func (r *ReconcileGitlab) reconcileDeployments(cr *gitlabv1beta1.Gitlab) error {
	gitlabCore := getGitlabDeployment(cr)

	if IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: gitlabCore.Name}, gitlabCore) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, gitlabCore, r.scheme); err != nil {
		return err
	}

	if err := r.client.Create(context.TODO(), gitlabCore); err != nil {
		return err
	}

	return nil
}

func (r *ReconcileGitlab) reconcileStatefulSets(cr *gitlabv1beta1.Gitlab) error {
	redis := getRedisStatefulSet(cr)

	if IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: redis.Name}, redis) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, redis, r.scheme); err != nil {
		return err
	}

	if err := r.client.Create(context.TODO(), redis); err != nil {
		return err
	}

	postgres := getPostgresStatefulSet(cr)

	if IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: postgres.Name}, postgres) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, postgres, r.scheme); err != nil {
		return err
	}

	if err := r.client.Create(context.TODO(), postgres); err != nil {
		return err
	}

	return nil
}

func (r *ReconcileGitlab) reconcileIngress(cr *gitlabv1beta1.Gitlab) error {
	ingress := getGitlabIngress(cr)

	if IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: ingress.Name}, ingress) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, ingress, r.scheme); err != nil {
		return err
	}

	if err := r.client.Create(context.TODO(), ingress); err != nil {
		return err
	}

	return nil
}

func (r *ReconcileGitlab) reconcileRoute(cr *gitlabv1beta1.Gitlab) error {
	workhorse := getGitlabRoute(cr)

	if IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: workhorse.Name}, workhorse) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, workhorse, r.scheme); err != nil {
		return err
	}

	if err := r.client.Create(context.TODO(), workhorse); err != nil {
		return err
	}

	registry := getRegistryRoute(cr)

	if IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: registry.Name}, registry) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, registry, r.scheme); err != nil {
		return err
	}

	if err := r.client.Create(context.TODO(), registry); err != nil {
		return err
	}

	return nil
}

func (r *ReconcileGitlab) reconcileServiceMonitor(cr *gitlabv1beta1.Gitlab) error {

	servicemon := getServiceMonitor(cr)

	if IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: servicemon.Name}, servicemon) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, servicemon, r.scheme); err != nil {
		return err
	}

	if err := r.client.Create(context.TODO(), servicemon); err != nil {
		return err
	}

	return nil
}
