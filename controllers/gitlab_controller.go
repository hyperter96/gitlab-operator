/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"

	certmanagerv1alpha2 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"

	nginxv1alpha1 "github.com/nginxinc/nginx-ingress-operator/pkg/apis/k8s/v1alpha1"
	routev1 "github.com/openshift/api/route/v1"
	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
	gitlabctl "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/gitlab"
	gitlabutils "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/utils"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	"k8s.io/apimachinery/pkg/api/errors"
)

// GitLabReconciler reconciles a GitLab object
type GitLabReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=apps.gitlab.com,resources=gitlabs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.gitlab.com,resources=gitlabs/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps.gitlab.com,resources=gitlabs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=endpoints,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=cronjobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=extensions,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=servicemonitors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=prometheuses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cert-manager.io,resources=issuers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cert-manager.io,resources=certificates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=route.openshift.io,resources=routes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=route.openshift.io,resources=routes/custom-host,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=k8s.nginx.org,resources=nginxingresscontrollers,verbs=get;list;watch;create;update;patch;delete

// Reconcile triggers when an event occurs on the watched resource
func (r *GitLabReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("gitlab", req.NamespacedName)

	log.Info("Reconciling GitLab", "name", req.NamespacedName.Name, "namespace", req.NamespacedName.Namespace)
	gitlab := &gitlabv1beta1.GitLab{}
	if err := r.Get(ctx, req.NamespacedName, gitlab); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		// could not get GitLab resource
		return ctrl.Result{}, err
	}

	if err := r.reconcileServiceAccount(ctx, gitlab); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileNamespaces(ctx); err != nil {
		return ctrl.Result{}, err
	}

	adapter := gitlabctl.NewCustomResourceAdapter(gitlab)

	if err := r.runSharedSecretsJob(adapter); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.runSelfSignedCertsJob(adapter); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileConfigMaps(gitlab); err != nil {
		return ctrl.Result{}, err
	}

	// if err := r.reconcileSecrets(gitlab); err != nil {
	// 	return ctrl.Result{}, err
	// }

	if err := r.reconcileServices(gitlab); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileStatefulSets(ctx, gitlab, log); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileMinioInstance(gitlab); err != nil {
		return ctrl.Result{}, err
	}

	if gitlabctl.RequiresCertManagerCertificate(gitlab).All() {
		if err := r.reconcileCertManagerCertificates(gitlab); err != nil {
			return ctrl.Result{}, err
		}
	}

	waitInterval := 5 * time.Second
	if !r.ifCoreServicesReady(ctx, gitlab) {
		return ctrl.Result{RequeueAfter: waitInterval}, nil
	}

	if err := r.reconcileJobs(gitlab); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileDeployments(ctx, gitlab); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.setupAutoscaling(ctx, gitlab); err != nil {
		return ctrl.Result{}, err
	}

	// Deploy route is on Openshift, Ingress otherwise
	if err := r.exposeGitLabInstance(gitlab); err != nil {
		return ctrl.Result{}, err
	}

	if gitlabutils.IsPrometheusSupported() {
		// Deploy a prometheus service monitor
		if err := r.reconcileServiceMonitor(gitlab); err != nil {
			return ctrl.Result{}, err
		}
	}

	if err := r.reconcileGitlabStatus(gitlab); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager configures the custom resource watched resources
func (r *GitLabReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gitlabv1beta1.GitLab{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&appsv1.Deployment{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&batchv1.Job{}).
		Owns(&extensionsv1beta1.Ingress{}).
		Owns(&monitoringv1.ServiceMonitor{}).
		Owns(&certmanagerv1alpha2.Issuer{}).
		Owns(&certmanagerv1alpha2.Certificate{}).
		Owns(&nginxv1alpha1.NginxIngressController{}).
		Complete(r)
}

func (r *GitLabReconciler) runSharedSecretsJob(adapter gitlabctl.CustomResourceAdapter) error {
	cfgMap, job, err := gitlabctl.SharedSecretsResources(adapter)
	if err != nil {
		return err
	}

	logger := r.Log.WithValues("gitlab", adapter.Reference(), "job", job.Name, "namespace", job.Namespace)

	logger.V(1).Info("Ensuring Job's ConfigMap exists", "configmap", cfgMap.Name)
	if err := r.createKubernetesResource(cfgMap, adapter.Resource()); err != nil {
		return err
	}

	return r.runJobAndWait(adapter, job)
}

func (r *GitLabReconciler) runSelfSignedCertsJob(adapter gitlabctl.CustomResourceAdapter) error {
	job, err := gitlabctl.SelfSignedCertsJob(adapter)
	if err != nil {
		return err
	}

	return r.runJobAndWait(adapter, job)
}

func (r *GitLabReconciler) runJobAndWait(adapter gitlabctl.CustomResourceAdapter, job *batchv1.Job) error {

	logger := r.Log.WithValues("gitlab", adapter.Reference(), "job", job.Name, "namespace", job.Namespace)

	lookupKey := types.NamespacedName{
		Name:      job.Name,
		Namespace: job.Namespace,
	}

	logger.V(1).Info("Creating Job")
	if err := r.createKubernetesResource(job, adapter.Resource()); err != nil {
		return err
	}

	logger.Info("Waiting for Job to finish")

	elapsed := time.Duration(0)
	timeout := gitlabctl.SharedSecretsJobTimeout()
	waitPeriod := gitlabctl.SharedSecretsJobWaitPeriod(timeout, elapsed)

	var result error = nil

	for {
		if elapsed > timeout {
			result = errors.NewTimeoutError("The Job did not finish in time", int(timeout))
			logger.Error(result, "Timeout for Job exceeded.",
				"timeout", timeout)
			break
		}

		logger.V(1).Info("Checking the status of Job")
		lookupVal := &batchv1.Job{}
		if err := r.Get(context.Background(), lookupKey, lookupVal); err != nil {
			logger.V(1).Info("Failed to check the status of Job. Skipping.", "error", err)

			/*
			 * This will make sure we won't stuck here forever,
			 * in case the error is recurring.
			 */
			clientDelay, _ := errors.SuggestsClientDelay(err)
			if clientDelay == 0 {
				clientDelay = 1
			}
			delay := time.Duration(clientDelay) * time.Second
			elapsed += delay
			time.Sleep(delay)

			continue
		}

		if lookupVal.Status.Succeeded > 0 {
			logger.Info("Success! The Job is complete.")
			break
		}

		if lookupVal.Status.Failed > 0 {
			result = errors.NewInternalError(
				fmt.Errorf("The %s Job has failed. Check the log output of the Job: %s", job.Name, lookupKey))
			logger.Error(result, "Failure! The Job is complete.")
			break
		}

		elapsed += waitPeriod
		time.Sleep(waitPeriod)
	}

	return result
}

//	Reconciler for all ConfigMaps come below
func (r *GitLabReconciler) reconcileConfigMaps(cr *gitlabv1beta1.GitLab) error {
	var configmaps []*corev1.ConfigMap

	/*
	 * TODO: reconcileShellDeployment must receive the adapter instead of
	 *       the CR itself and the following line should be removed.
	 */
	adapter := gitlabctl.NewCustomResourceAdapter(cr)

	shell := gitlabctl.ShellConfigMaps(adapter)
	taskRunner := gitlabctl.TaskRunnerConfigMap(adapter)
	gitaly := gitlabctl.GitalyConfigMap(adapter)
	exporter := gitlabctl.ExporterConfigMaps(adapter)
	webservice := gitlabctl.WebserviceConfigMaps(adapter)
	migration := gitlabctl.MigrationsConfigMap(adapter)
	sidekiq := gitlabctl.SidekiqConfigMaps(adapter)
	redis := gitlabctl.RedisConfigMaps(adapter)
	postgres := gitlabctl.PostgresConfigMap(adapter)

	workhorse := gitlabctl.WorkhorseConfigMap(cr)

	gitlab := gitlabctl.GetGitLabConfigMap(cr)

	registry := gitlabctl.RegistryConfigMap(cr)

	configmaps = append(configmaps,
		gitaly,
		workhorse,
		gitlab,
		registry,
		taskRunner,
		migration,
		postgres,
	)
	configmaps = append(configmaps, shell...)
	configmaps = append(configmaps, exporter...)
	configmaps = append(configmaps, webservice...)
	configmaps = append(configmaps, sidekiq...)
	configmaps = append(configmaps, redis...)

	for _, cm := range configmaps {
		if err := r.createKubernetesResource(cm, cr); err != nil {
			return err
		}
	}

	return nil
}

func (r *GitLabReconciler) reconcileJobs(cr *gitlabv1beta1.GitLab) error {

	// initialize buckets once s3 storage is up
	buckets := gitlabctl.BucketCreationJob(cr)
	if err := r.createKubernetesResource(buckets, cr); err != nil {
		return err
	}

	// migration := gitlabctl.MigrationsJob(cr)
	// return r.createKubernetesResource(migration, cr)

	return r.runMigrationsJob(cr)
}

func (r *GitLabReconciler) reconcileServiceMonitor(cr *gitlabv1beta1.GitLab) error {
	var servicemonitors []*monitoringv1.ServiceMonitor

	gitaly := gitlabctl.GitalyServiceMonitor(cr)

	gitlab := gitlabctl.ExporterServiceMonitor(cr)

	postgres := gitlabctl.PostgresqlServiceMonitor(cr)

	redis := gitlabctl.RedisServiceMonitor(cr)

	workhorse := gitlabctl.WebserviceServiceMonitor(cr)

	servicemonitors = append(servicemonitors,
		gitlab,
		gitaly,
		postgres,
		redis,
		workhorse,
	)

	for _, sm := range servicemonitors {
		if err := r.createKubernetesResource(sm, cr); err != nil {
			return err
		}
	}

	service := gitlabctl.ExposePrometheusCluster(cr)
	if err := r.createKubernetesResource(service, nil); err != nil {
		return err
	}

	prometheus := gitlabctl.PrometheusCluster(cr)
	return r.createKubernetesResource(prometheus, nil)
}

func (r *GitLabReconciler) runMigrationsJob(cr *gitlabv1beta1.GitLab) error {
	/*
	 * TODO: runMigrationsJob must receive the adapter instead of
	 *       the CR itself and the following line should be removed.
	 */
	adapter := gitlabctl.NewCustomResourceAdapter(cr)

	migrations, err := gitlabctl.MigrationsJob(adapter)
	if err != nil {
		return err
	}

	lookupKey := types.NamespacedName{
		Name:      migrations.Name,
		Namespace: migrations.Namespace,
	}

	logger := r.Log.WithValues("gitlab", adapter.Reference(), "job", lookupKey)
	logger.V(1).Info("Creating migrations Job", "name", migrations.Name)
	if err := r.createKubernetesResource(migrations, cr); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcileDeployments(ctx context.Context, cr *gitlabv1beta1.GitLab) error {

	if err := r.reconcileWebserviceDeployment(ctx, cr); err != nil {
		return err
	}

	if err := r.reconcileShellDeployment(ctx, cr); err != nil {
		return err
	}

	if err := r.reconcileSidekiqDeployment(ctx, cr); err != nil {
		return err
	}

	if err := r.reconcileRegistryDeployment(ctx, cr); err != nil {
		return err
	}

	if err := r.reconcileTaskRunnerDeployment(ctx, cr); err != nil {
		return err
	}

	if err := r.reconcileGitlabExporterDeployment(ctx, cr); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcileStatefulSets(ctx context.Context, cr *gitlabv1beta1.GitLab, log logr.Logger) error {

	var statefulsets []*appsv1.StatefulSet

	/*
	 * TODO: reconcileShellDeployment must receive the adapter instead of
	 *       the CR itself and the following line should be removed.
	 */
	adapter := gitlabctl.NewCustomResourceAdapter(cr)

	gitaly := gitlabctl.GitalyStatefulSet(adapter)
	redis := gitlabctl.RedisStatefulSet(adapter)
	postgres := gitlabctl.PostgresStatefulSet(adapter)

	statefulsets = append(statefulsets, postgres, redis, gitaly)

	for _, statefulset := range statefulsets {
		if err := r.createKubernetesResource(statefulset, cr); err != nil {
			return err
		}
	}

	return nil
}

func (r *GitLabReconciler) createKubernetesResource(object interface{}, parent *gitlabv1beta1.GitLab) error {

	if r.isObjectFound(object) {
		return nil
	}

	// If parent resource is nil, not owner reference will be set
	if parent != nil {
		if err := controllerutil.SetControllerReference(parent, object.(metav1.Object), r.Scheme); err != nil {
			return err
		}
	}

	return r.Create(context.TODO(), object.(runtime.Object).DeepCopyObject())
}

func (r *GitLabReconciler) maskEmailPasword(cr *gitlabv1beta1.GitLab) error {
	gitlab := &gitlabv1beta1.GitLab{}
	r.Get(context.TODO(), types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, gitlab)

	// If password is stored in secret and is still visible in CR, update it to emty string
	emailPasswd, err := gitlabutils.GetSecretValue(r.Client, cr.Namespace, cr.Name+"-smtp-settings-secret", "smtp_user_password")
	if err != nil {
		// log.Error(err, "")
	}

	if gitlab.Spec.SMTP.Password == emailPasswd && cr.Spec.SMTP.Password != "" {
		// Update CR
		gitlab.Spec.SMTP.Password = ""
		if err := r.Update(context.TODO(), gitlab); err != nil && errors.IsResourceExpired(err) {
			return err
		}
	}

	// If stored password does not match the CR password,
	// update the secret and empty the password string in Gitlab CR

	return nil
}

func (r *GitLabReconciler) reconcileMinioInstance(cr *gitlabv1beta1.GitLab) error {
	cm := gitlabctl.MinioScriptConfigMap(cr)
	if err := r.createKubernetesResource(cm, cr); err != nil {
		return err
	}

	secret := gitlabctl.MinioSecret(cr)
	if err := r.createKubernetesResource(secret, cr); err != nil && errors.IsAlreadyExists(err) {
		return err
	}

	// Only deploy the minio service and statefulset for development builds
	if cr.Spec.ObjectStore.Development {
		svc := gitlabctl.MinioService(cr)
		if err := r.createKubernetesResource(svc, cr); err != nil {
			return err
		}

		// deploy minio
		minio := gitlabctl.MinioStatefulSet(cr)
		return r.createKubernetesResource(minio, cr)
	}

	return nil
}

// func (r *GitLabReconciler) reconcileSecrets(cr *gitlabv1beta1.GitLab) error {
// 	var secrets []*corev1.Secret

// 	gitaly := gitlabctl.GitalySecret(cr)

// 	workhorse := gitlabctl.WorkhorseSecret(cr)

// 	registry := gitlabctl.RegistryHTTPSecret(cr)

// 	registryCert := gitlabctl.RegistryCertSecret(cr)

// 	rails := gitlabctl.RailsSecret(cr)

// 	postgres := gitlabctl.PostgresSecret(cr)

// 	redis := gitlabctl.RedisSecret(cr)

// 	runner := gitlabctl.RunnerRegistrationSecret(cr)

// 	root := gitlabctl.RootUserSecret(cr)

// 	smtp := gitlabctl.SMTPSettingsSecret(cr)

// 	shell := gitlabctl.ShellSecret(cr)

// 	keys := gitlabctl.ShellSSHKeysSecret(cr)

// 	secrets = append(secrets,
// 		gitaly,
// 		registry,
// 		registryCert,
// 		workhorse,
// 		rails,
// 		postgres,
// 		redis,
// 		root,
// 		runner,
// 		smtp,
// 		shell,
// 		keys,
// 	)

// 	for _, secret := range secrets {
// 		if err := r.createKubernetesResource(secret, cr); err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

func (r *GitLabReconciler) reconcileServices(cr *gitlabv1beta1.GitLab) error {
	var services []*corev1.Service

	/*
	 * TODO: reconcileShellDeployment must receive the adapter instead of
	 *       the CR itself and the following line should be removed.
	 */
	adapter := gitlabctl.NewCustomResourceAdapter(cr)

	shell := gitlabctl.ShellService(adapter)
	gitaly := gitlabctl.GitalyService(adapter)
	exporter := gitlabctl.ExporterService(adapter)
	webservice := gitlabctl.WebserviceService(adapter)
	redis := gitlabctl.RedisServices(adapter)
	postgres := gitlabctl.PostgresServices(adapter)

	registry := gitlabctl.RegistryService(cr)

	services = append(services,
		gitaly,
		registry,
		webservice,
		shell,
		exporter,
	)
	services = append(services, redis...)
	services = append(services, postgres...)

	for _, svc := range services {
		if err := r.createKubernetesResource(svc, cr); err != nil {
			return err
		}
	}

	return nil
}

func (r *GitLabReconciler) reconcileGitlabExporterDeployment(ctx context.Context, cr *gitlabv1beta1.GitLab) error {
	/*
	 * TODO: reconcileExporterDeployment must receive the adapter instead of
	 *       the CR itself and the following line should be removed.
	 */
	adapter := gitlabctl.NewCustomResourceAdapter(cr)

	exporter := gitlabctl.ExporterDeployment(adapter)

	if err := controllerutil.SetControllerReference(cr, exporter, r.Scheme); err != nil {
		return err
	}

	found := &appsv1.Deployment{}
	lookupKey := types.NamespacedName{
		Name:      exporter.Name,
		Namespace: exporter.Namespace,
	}
	if err := r.Get(ctx, lookupKey, found); err != nil {
		if errors.IsNotFound(err) {
			return r.Create(ctx, exporter.DeepCopy())
		}

		return err
	}

	deployment, changed := gitlabutils.IsDeploymentChanged(found, exporter.DeepCopy())
	if changed {
		return r.Update(ctx, deployment)
	}

	return nil
}

func (r *GitLabReconciler) reconcileWebserviceDeployment(ctx context.Context, cr *gitlabv1beta1.GitLab) error {

	/*
	 * TODO: reconcileShellDeployment must receive the adapter instead of
	 *       the CR itself and the following line should be removed.
	 */
	adapter := gitlabctl.NewCustomResourceAdapter(cr)

	webservice := gitlabctl.WebserviceDeployment(adapter)

	if err := controllerutil.SetControllerReference(cr, webservice, r.Scheme); err != nil {
		return err
	}

	found := &appsv1.Deployment{}
	lookupKey := types.NamespacedName{
		Name:      webservice.Name,
		Namespace: webservice.Namespace,
	}
	if err := r.Get(ctx, lookupKey, found); err != nil {
		if errors.IsNotFound(err) {
			return r.Create(ctx, webservice.DeepCopy())
		}

		return err
	}

	deployment, changed := gitlabutils.IsDeploymentChanged(found, webservice.DeepCopy())
	if changed {
		return r.Update(ctx, deployment)
	}

	return nil
}

func (r *GitLabReconciler) reconcileRegistryDeployment(ctx context.Context, cr *gitlabv1beta1.GitLab) error {
	registry := gitlabctl.RegistryDeployment(cr)

	if err := controllerutil.SetControllerReference(cr, registry, r.Scheme); err != nil {
		return err
	}

	found := &appsv1.Deployment{}
	lookupKey := types.NamespacedName{
		Name:      registry.Name,
		Namespace: registry.Namespace,
	}
	if err := r.Get(ctx, lookupKey, found); err != nil {
		if errors.IsNotFound(err) {
			return r.Create(ctx, registry.DeepCopy())
		}

		return err
	}

	deployment, changed := gitlabutils.IsDeploymentChanged(found, registry.DeepCopy())
	if changed {
		return r.Update(ctx, deployment)
	}

	return nil
}

func (r *GitLabReconciler) reconcileShellDeployment(ctx context.Context, cr *gitlabv1beta1.GitLab) error {

	/*
	 * TODO: reconcileShellDeployment must receive the adapter instead of
	 *       the CR itself and the following line should be removed.
	 */
	adapter := gitlabctl.NewCustomResourceAdapter(cr)

	shell := gitlabctl.ShellDeployment(adapter)

	if err := controllerutil.SetControllerReference(cr, shell, r.Scheme); err != nil {
		return err
	}

	found := &appsv1.Deployment{}
	lookupKey := types.NamespacedName{
		Name:      shell.Name,
		Namespace: shell.Namespace,
	}
	if err := r.Get(ctx, lookupKey, found); err != nil {
		if errors.IsNotFound(err) {
			return r.Create(ctx, shell)
		}

		return err
	}

	deployment, changed := gitlabutils.IsDeploymentChanged(found, shell.DeepCopy())
	if changed {
		return r.Update(ctx, deployment)
	}

	return nil
}

func (r *GitLabReconciler) reconcileSidekiqDeployment(ctx context.Context, cr *gitlabv1beta1.GitLab) error {
	/*
	 * TODO: reconcileShellDeployment must receive the adapter instead of
	 *       the CR itself and the following line should be removed.
	 */
	adapter := gitlabctl.NewCustomResourceAdapter(cr)

	sidekiq := gitlabctl.SidekiqDeployment(adapter)

	if err := controllerutil.SetControllerReference(cr, sidekiq, r.Scheme); err != nil {
		return err
	}

	found := &appsv1.Deployment{}
	lookupKey := types.NamespacedName{
		Name:      sidekiq.Name,
		Namespace: sidekiq.Namespace,
	}
	if err := r.Get(ctx, lookupKey, found); err != nil {
		if errors.IsNotFound(err) {
			return r.Create(ctx, sidekiq.DeepCopy())
		}

		return err
	}

	deployment, changed := gitlabutils.IsDeploymentChanged(found, sidekiq.DeepCopy())
	if changed {
		return r.Update(ctx, deployment)
	}

	return nil
}

func (r *GitLabReconciler) reconcileTaskRunnerDeployment(ctx context.Context, cr *gitlabv1beta1.GitLab) error {
	/*
	 * TODO: reconcileTaskRunnerDeployment must receive the adapter instead of
	 *       the CR itself and the following line should be removed.
	 */
	adapter := gitlabctl.NewCustomResourceAdapter(cr)

	tasker := gitlabctl.TaskRunnerDeployment(adapter)

	if err := controllerutil.SetControllerReference(cr, tasker, r.Scheme); err != nil {
		return err
	}

	found := &appsv1.Deployment{}
	lookupKey := types.NamespacedName{
		Name:      tasker.Name,
		Namespace: tasker.Namespace,
	}
	if err := r.Get(ctx, lookupKey, found); err != nil {
		if errors.IsNotFound(err) {
			return r.Create(ctx, tasker.DeepCopy())
		}

		return err
	}

	deployment, changed := gitlabutils.IsDeploymentChanged(found, tasker.DeepCopy())
	if changed {
		return r.Update(ctx, deployment)
	}

	return nil
}

func (r *GitLabReconciler) exposeGitLabInstance(cr *gitlabv1beta1.GitLab) error {
	// if gitlabutils.IsOpenshift() {
	// 	return r.reconcileRoute(cr)
	// }

	return r.reconcileIngress(cr)
}

func (r *GitLabReconciler) reconcileRoute(cr *gitlabv1beta1.GitLab) error {
	app := gitlabctl.MainRoute(cr)

	admin := gitlabctl.AdminRoute(cr)

	registry := gitlabctl.RegistryRoute(cr)

	var routes []*routev1.Route
	routes = append(routes,
		app,
		admin,
		registry,
	)

	for _, route := range routes {
		if err := r.createKubernetesResource(route, cr); err != nil {
			return err
		}
	}

	return nil
}

func (r *GitLabReconciler) reconcileIngress(cr *gitlabv1beta1.GitLab) error {

	controller := gitlabctl.IngressController(cr)
	if err := r.createKubernetesResource(controller, cr); err != nil {
		return err
	}

	var ingresses []*extensionsv1beta1.Ingress
	gitlab := gitlabctl.Ingress(cr)

	registry := gitlabctl.RegistryIngress(cr)

	ingresses = append(ingresses,
		gitlab,
		registry,
	)

	for _, ingress := range ingresses {
		if err := r.createKubernetesResource(ingress, cr); err != nil {
			return err
		}
	}

	return nil
}

func (r *GitLabReconciler) reconcileCertManagerCertificates(cr *gitlabv1beta1.GitLab) error {
	// certificates := RequiresCertificate(cr)

	issuer := gitlabctl.CertificateIssuer(cr)

	return r.createKubernetesResource(issuer, cr)
}

func (r *GitLabReconciler) setupAutoscaling(ctx context.Context, cr *gitlabv1beta1.GitLab) error {
	selector := client.MatchingLabelsSelector{
		Selector: getLabelSet(cr).AsSelector(),
	}

	deployments := &appsv1.DeploymentList{}
	err := r.List(ctx, deployments, client.InNamespace(cr.Namespace), selector)
	if err != nil {
		return err
	}

	for _, deploy := range deployments.Items {
		if err := r.reconcileHPA(ctx, &deploy, cr); err != nil {
			return err
		}
	}

	return nil
}

func (r *GitLabReconciler) reconcileHPA(ctx context.Context, deployment *appsv1.Deployment, cr *gitlabv1beta1.GitLab) error {
	excludedDeployments := [2]string{"gitlab-exporter", "gitlab-task-runner"}
	for _, excludedDeployment := range excludedDeployments {
		if strings.Contains(deployment.Name, excludedDeployment) {
			return nil
		}
	}

	hpa := gitlabctl.HorizontalAutoscaler(deployment, cr)

	found := &autoscalingv1.HorizontalPodAutoscaler{}
	err := r.Get(ctx, types.NamespacedName{Name: deployment.Name, Namespace: cr.Namespace}, found)
	if err != nil {
		if errors.IsNotFound(err) {
			// return nil if hpa is nil
			if hpa == nil {
				return nil
			}

			if err := controllerutil.SetControllerReference(cr, hpa, r.Scheme); err != nil {
				return err
			}

			return r.Create(ctx, hpa.DeepCopy())
		}

		return err
	}

	if cr.Spec.AutoScaling == nil {
		return r.Delete(ctx, found)
	}

	if !reflect.DeepEqual(hpa.Spec, found.Spec) {
		if *found.Spec.MinReplicas != *hpa.Spec.MinReplicas {
			found.Spec.MinReplicas = hpa.Spec.MinReplicas
		}

		if found.Spec.MaxReplicas != hpa.Spec.MaxReplicas {
			found.Spec.MaxReplicas = hpa.Spec.MaxReplicas
		}

		if found.Spec.TargetCPUUtilizationPercentage != hpa.Spec.TargetCPUUtilizationPercentage {
			found.Spec.TargetCPUUtilizationPercentage = hpa.Spec.TargetCPUUtilizationPercentage
		}

		return r.Update(ctx, found)
	}

	return nil
}

func (r *GitLabReconciler) reconcileNamespaces(ctx context.Context) error {
	secured := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "gitlab-secured-apps",
		},
	}

	if err := r.createNamespace(ctx, secured); err != nil {
		return err
	}

	managed := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "gitlab-managed-apps",
		},
	}

	return r.createNamespace(ctx, managed)
}

func (r *GitLabReconciler) createNamespace(ctx context.Context, namespace *corev1.Namespace) error {
	found := &corev1.Namespace{}
	err := r.Get(ctx, types.NamespacedName{Name: namespace.Name}, found)
	if err != nil {
		// create namespace if doesnt exist
		if errors.IsNotFound(err) {
			return r.Create(ctx, namespace)
		}

		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcileServiceAccount(ctx context.Context, cr *gitlabv1beta1.GitLab) error {
	sa := gitlabutils.ServiceAccount("gitlab-app", cr.Namespace)

	found := &corev1.ServiceAccount{}
	lookupKey := types.NamespacedName{Name: sa.Name, Namespace: cr.Namespace}
	if err := r.Get(ctx, lookupKey, found); err != nil {
		if errors.IsNotFound(err) {
			if err := r.Create(ctx, sa); err != nil {
				return err
			}

			return nil
		}

		return err
	}

	return nil
}
