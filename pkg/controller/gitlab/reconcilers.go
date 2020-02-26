package gitlab

import (
	"context"

	gitlabv1beta1 "github.com/OchiengEd/gitlab-operator/pkg/apis/gitlab/v1beta1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *ReconcileGitlab) reconcileConfigMaps(cr *gitlabv1beta1.Gitlab, s security) error {
	gitlabConf := getGitlabConfig(cr)

	if r.isObjectFound(types.NamespacedName{Name: gitlabConf.Name, Namespace: gitlabConf.Namespace}, gitlabConf) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, gitlabConf, r.scheme); err != nil {
		return err
	}

	if err := r.client.Create(context.TODO(), gitlabConf); err != nil {
		return err
	}

	redis := getRedisConfig(cr, s)

	if r.isObjectFound(types.NamespacedName{Name: redis.Name, Namespace: redis.Namespace}, redis) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, redis, r.scheme); err != nil {
		return err
	}

	if err := r.client.Create(context.TODO(), redis); err != nil {
		return err
	}

	postgresInit := getPostgresInitdbConfig(cr)

	if r.isObjectFound(types.NamespacedName{Name: postgresInit.Name, Namespace: postgresInit.Namespace}, postgresInit) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, postgresInit, r.scheme); err != nil {
		return err
	}

	if err := r.client.Create(context.TODO(), postgresInit); err != nil {
		return err
	}

	return nil
}

func (r *ReconcileGitlab) reconcileSecrets(cr *gitlabv1beta1.Gitlab, s security) error {

	core := getGilabSecret(cr, s)

	if r.isObjectFound(types.NamespacedName{Name: core.Name, Namespace: core.Namespace}, core) {
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

	if r.isObjectFound(types.NamespacedName{Name: postgres.Name, Namespace: postgres.Namespace}, postgres) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, postgres, r.scheme); err != nil {
		return err
	}

	if err := r.client.Create(context.TODO(), postgres); err != nil {
		return err
	}

	redis := getRedisService(cr)

	if r.isObjectFound(types.NamespacedName{Name: redis.Name, Namespace: redis.Namespace}, redis) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, redis, r.scheme); err != nil {
		return err
	}

	if err := r.client.Create(context.TODO(), redis); err != nil {
		return err
	}

	gitlab := getGitlabService(cr)

	if r.isObjectFound(types.NamespacedName{Name: gitlab.Name, Namespace: gitlab.Namespace}, gitlab) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, gitlab, r.scheme); err != nil {
		return err
	}

	if err := r.client.Create(context.TODO(), gitlab); err != nil {
		return err
	}

	return nil
}

func (r *ReconcileGitlab) reconcilePersistentVolumeClaims(cr *gitlabv1beta1.Gitlab) error {
	if cr.Spec.Registry.Enabled {
		registryVolume := getRegistryVolumeClaim(cr)

		if r.isObjectFound(types.NamespacedName{Name: registryVolume.Name, Namespace: registryVolume.Namespace}, registryVolume) {
			return nil
		}

		if err := controllerutil.SetControllerReference(cr, registryVolume, r.scheme); err != nil {
			return err
		}

		if err := r.client.Create(context.TODO(), registryVolume); err != nil {
			return err
		}
	}

	dataVolume := getGitlabDataVolumeClaim(cr)

	if r.isObjectFound(types.NamespacedName{Name: dataVolume.Name, Namespace: dataVolume.Namespace}, dataVolume) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, dataVolume, r.scheme); err != nil {
		return err
	}

	if err := r.client.Create(context.TODO(), dataVolume); err != nil {
		return err
	}

	configVolume := getGitlabConfigVolumeClaim(cr)

	if r.isObjectFound(types.NamespacedName{Name: configVolume.Name, Namespace: configVolume.Namespace}, configVolume) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, configVolume, r.scheme); err != nil {
		return err
	}

	if err := r.client.Create(context.TODO(), configVolume); err != nil {
		return err
	}

	return nil
}

func (r *ReconcileGitlab) reconcileDeployments(cr *gitlabv1beta1.Gitlab) error {

	gitlabCore := getGitlabDeployment(cr)

	if r.isObjectFound(types.NamespacedName{Name: gitlabCore.Name, Namespace: gitlabCore.Namespace}, gitlabCore) {
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

	if r.isObjectFound(types.NamespacedName{Name: redis.Name, Namespace: redis.Namespace}, redis) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, redis, r.scheme); err != nil {
		return err
	}

	if err := r.client.Create(context.TODO(), redis); err != nil {
		return err
	}

	postgres := getPostgresStatefulSet(cr)

	if r.isObjectFound(types.NamespacedName{Name: postgres.Name, Namespace: postgres.Namespace}, postgres) {
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

	if r.isObjectFound(types.NamespacedName{Name: ingress.Name, Namespace: ingress.Namespace}, ingress) {
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
	route := getGitlabRoute(cr)

	if r.isObjectFound(types.NamespacedName{Name: route.Name, Namespace: route.Namespace}, route) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, route, r.scheme); err != nil {
		return err
	}

	if err := r.client.Create(context.TODO(), route); err != nil {
		return err
	}

	return nil
}
