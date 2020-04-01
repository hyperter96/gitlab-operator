package gitlab

import (
	"context"

	gitlabv1beta1 "gitlab.com/ochienged/gitlab-operator/pkg/apis/gitlab/v1beta1"
	gitlabutils "gitlab.com/ochienged/gitlab-operator/pkg/controller/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *ReconcileGitlab) reconcileConfigMaps(cr *gitlabv1beta1.Gitlab) error {

	if err := r.reconcileGitalyConfigMap(cr); err != nil {
		return err
	}

	if err := r.reconcileRedisConfigMap(cr); err != nil {
		return err
	}

	if err := r.reconcileUnicornConfigMap(cr); err != nil {
		return err
	}

	if err := r.reconcileWorkhorseConfigMap(cr); err != nil {
		return err
	}

	if err := r.reconcileGitlabConfigMap(cr); err != nil {
		return err
	}

	if err := r.reconcileShellConfigMap(cr); err != nil {
		return err
	}

	if err := r.reconcileSidekiqConfigMap(cr); err != nil {
		return err
	}

	if err := r.reconcileGitlabExporterConfigMap(cr); err != nil {
		return err
	}

	if err := r.reconcileRegistryConfigMap(cr); err != nil {
		return err
	}

	if err := r.reconcileTaskRunnerConfigMap(cr); err != nil {
		return err
	}

	if err := r.reconcileSidekiqConfigMap(cr); err != nil {
		return err
	}

	if err := r.reconcileMigrationsConfigMap(cr); err != nil {
		return err
	}

	if err := r.reconcilePostgresInitDBConfigMap(cr); err != nil {
		return err
	}

	if err := r.reconcileRedisScriptsConfigMap(cr); err != nil {
		return err
	}

	return nil
}

func (r *ReconcileGitlab) reconcileSecrets(cr *gitlabv1beta1.Gitlab) error {

	if err := r.reconcileGitlabSecret(cr); err != nil {
		return err
	}

	if err := r.reconcilePostgresSecret(cr); err != nil {
		return err
	}

	if err := r.reconcileRedisSecret(cr); err != nil {
		return err
	}

	if err := r.reconcileShellSSHKeysSecret(cr); err != nil {
		return err
	}

	if err := r.reconcileShellSecret(cr); err != nil {
		return err
	}

	if err := r.reconcileRegistrySecret(cr); err != nil {
		return err
	}

	if err := r.reconcileWorkhorseSecret(cr); err != nil {
		return err
	}

	if err := r.reconcileGitalySecret(cr); err != nil {
		return err
	}

	if err := r.reconcileRegistryHTTPSecret(cr); err != nil {
		return err
	}

	if err := r.reconcileRailsSecret(cr); err != nil {
		return err
	}

	if err := r.reconcileMinioSecret(cr); err != nil {
		return err
	}

	return nil
}

func (r *ReconcileGitlab) maskEmailPasword(cr *gitlabv1beta1.Gitlab) error {
	gitlab := &gitlabv1beta1.Gitlab{}
	r.client.Get(context.TODO(), types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, gitlab)

	if gitlab.Spec.Redis.Replicas == 0 {
		gitlab.Spec.Redis.Replicas = 1
	}

	if gitlab.Spec.Database.Replicas == 0 {
		gitlab.Spec.Database.Replicas = 1
	}

	// If password is stored in secret and is still visible in CR, update it to emty string
	emailPasswd := gitlabutils.GetSecretValue(r.client, cr.Namespace, cr.Name+"-gitlab-secrets", "smtp_user_password")
	if gitlab.Spec.SMTP.Password == emailPasswd && cr.Spec.SMTP.Password != "" {
		// Update CR
		gitlab.Spec.SMTP.Password = ""
		if err := r.client.Update(context.TODO(), gitlab); err != nil && errors.IsResourceExpired(err) {
			return err
		}
	}

	// If stored password does not match the CR password,
	// update the secret and empty the password string in Gitlab CR

	return nil
}

func (r *ReconcileGitlab) reconcileJobs(cr *gitlabv1beta1.Gitlab) error {

	if err := r.reconcileMigrationsJob(cr); err != nil {
		return err
	}

	return nil
}

func (r *ReconcileGitlab) reconcileServices(cr *gitlabv1beta1.Gitlab) error {
	if err := r.reconcilePostgresService(cr); err != nil {
		return err
	}

	if err := r.reconcilePostgresHeadlessService(cr); err != nil {
		return err
	}

	if err := r.reconcileRedisService(cr); err != nil {
		return err
	}

	if err := r.reconcileRedisHeadlessService(cr); err != nil {
		return err
	}

	if err := r.reconcileGitalyService(cr); err != nil {
		return err
	}

	if err := r.reconcileRegistryService(cr); err != nil {
		return err
	}

	if err := r.reconcileUnicornService(cr); err != nil {
		return err
	}

	if err := r.reconcileShellService(cr); err != nil {
		return err
	}

	if err := r.reconcileGitlabExporterService(cr); err != nil {
		return err
	}

	if err := r.reconcilePostgresMetricsService(cr); err != nil {
		return err
	}

	if err := r.reconcileRedisMetricsService(cr); err != nil {
		return err
	}

	return nil
}

func (r *ReconcileGitlab) reconcilePersistentVolumeClaims(cr *gitlabv1beta1.Gitlab) error {

	if cr.Spec.Registry.Enabled && cr.Spec.Volumes.Registry.Capacity != "" {
		registryVolume := getRegistryVolumeClaim(cr)

		if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: registryVolume.Name}, registryVolume) {
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

		if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: dataVolume.Name}, dataVolume) {
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

		if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: configVolume.Name}, configVolume) {
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

	if err := r.reconcileUnicornDeployment(cr); err != nil {
		return err
	}

	if err := r.reconcileShellDeployment(cr); err != nil {
		return err
	}

	if err := r.reconcileSidekiqDeployment(cr); err != nil {
		return err
	}

	if err := r.reconcileRegistryDeployment(cr); err != nil {
		return err
	}

	if err := r.reconcileTaskRunnerDeployment(cr); err != nil {
		return err
	}

	if err := r.reconcileGitlabExporterDeployment(cr); err != nil {
		return err
	}

	return nil
}

func (r *ReconcileGitlab) reconcileStatefulSets(cr *gitlabv1beta1.Gitlab) error {

	if err := r.reconcileRedisStatefulSet(cr); err != nil {
		return err
	}

	if err := r.reconcilePostgresStatefulSet(cr); err != nil {
		return err
	}

	if err := r.reconcileGitalyStatefulSet(cr); err != nil {
		return err
	}

	return nil
}

func (r *ReconcileGitlab) reconcileIngress(cr *gitlabv1beta1.Gitlab) error {
	ingress := getGitlabIngress(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: ingress.Name}, ingress) {
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

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: workhorse.Name}, workhorse) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, workhorse, r.scheme); err != nil {
		return err
	}

	if err := r.client.Create(context.TODO(), workhorse); err != nil {
		return err
	}

	registry := getRegistryRoute(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: registry.Name}, registry) {
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

	if err := r.reconcilePrometheusServiceMonitor(cr); err != nil {
		return err
	}

	return nil
}
