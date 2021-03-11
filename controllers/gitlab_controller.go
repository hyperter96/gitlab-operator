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
	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/helpers"
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

	adapter := helpers.NewCustomResourceAdapter(gitlab)

	if err := r.reconcileServiceAccount(ctx, adapter); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileNamespaces(ctx); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.runSharedSecretsJob(ctx, adapter); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.runSelfSignedCertsJob(ctx, adapter); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileConfigMaps(ctx, adapter); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileServices(ctx, adapter); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileStatefulSets(ctx, adapter); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileMinioInstance(ctx, adapter); err != nil {
		return ctrl.Result{}, err
	}

	if gitlabctl.RequiresCertManagerCertificate(gitlab).All() {
		if err := r.reconcileCertManagerCertificates(ctx, adapter); err != nil {
			return ctrl.Result{}, err
		}
	}

	waitInterval := 5 * time.Second
	if !r.ifCoreServicesReady(ctx, adapter) {
		return ctrl.Result{RequeueAfter: waitInterval}, nil
	}

	if err := r.reconcileJobs(ctx, adapter); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileDeployments(ctx, adapter); err != nil {
		return ctrl.Result{}, err
	}

	// Disables autoscaling so the Operator does not attempt to
	// remove replicas that it is not expecting. Considered a temporary
	// fix until HPAs can be disabled in the Chart, and/or the Operator
	// is updated to accept replicas created by HPAs.
	// if err := r.setupAutoscaling(ctx, adapter); err != nil {
	//   return ctrl.Result{}, err
	// }

	// Deploy route is on Openshift, Ingress otherwise
	if err := r.exposeGitLabInstance(ctx, adapter); err != nil {
		return ctrl.Result{}, err
	}

	if gitlabutils.IsPrometheusSupported() {
		// Deploy a prometheus service monitor
		if err := r.reconcileServiceMonitor(ctx, adapter); err != nil {
			return ctrl.Result{}, err
		}
	}

	if err := r.reconcileGitlabStatus(ctx, adapter); err != nil {
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

func (r *GitLabReconciler) runSharedSecretsJob(ctx context.Context, adapter helpers.CustomResourceAdapter) error {
	cfgMap, job, err := gitlabctl.SharedSecretsResources(adapter)
	if err != nil {
		return err
	}

	logger := r.Log.WithValues("gitlab", adapter.Reference(), "job", job.Name, "namespace", job.Namespace)

	logger.V(1).Info("Ensuring Job's ConfigMap exists", "configmap", cfgMap.Name)
	if err := r.createKubernetesResource(ctx, cfgMap, adapter); err != nil {
		return err
	}

	return r.runJobAndWait(ctx, adapter, job)
}

func (r *GitLabReconciler) runSelfSignedCertsJob(ctx context.Context, adapter helpers.CustomResourceAdapter) error {
	job, err := gitlabctl.SelfSignedCertsJob(adapter)
	if err != nil {
		return err
	}

	return r.runJobAndWait(ctx, adapter, job)
}

func (r *GitLabReconciler) runJobAndWait(ctx context.Context, adapter helpers.CustomResourceAdapter, job *batchv1.Job) error {

	logger := r.Log.WithValues("gitlab", adapter.Reference(), "job", job.Name, "namespace", job.Namespace)

	lookupKey := types.NamespacedName{
		Name:      job.Name,
		Namespace: job.Namespace,
	}

	logger.V(1).Info("Creating Job")
	if err := r.createKubernetesResource(ctx, job, adapter); err != nil {
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
func (r *GitLabReconciler) reconcileConfigMaps(ctx context.Context, adapter helpers.CustomResourceAdapter) error {
	var configmaps []*corev1.ConfigMap

	shell := gitlabctl.ShellConfigMaps(adapter)
	taskRunner := gitlabctl.TaskRunnerConfigMap(adapter)
	gitaly := gitlabctl.GitalyConfigMap(adapter)
	exporter := gitlabctl.ExporterConfigMaps(adapter)
	webservice := gitlabctl.WebserviceConfigMaps(adapter)
	migration := gitlabctl.MigrationsConfigMap(adapter)
	sidekiq := gitlabctl.SidekiqConfigMaps(adapter)
	redis := gitlabctl.RedisConfigMaps(adapter)
	postgres := gitlabctl.PostgresConfigMap(adapter)
	registry := gitlabctl.RegistryConfigMap(adapter)

	gitlab := gitlabctl.GetGitLabConfigMap(adapter.Resource())

	configmaps = append(configmaps,
		gitaly,
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
		if err := r.createKubernetesResource(ctx, cm, adapter); err != nil {
			return err
		}
	}

	return nil
}

func (r *GitLabReconciler) reconcileJobs(ctx context.Context, adapter helpers.CustomResourceAdapter) error {

	// initialize buckets once s3 storage is up
	buckets := gitlabctl.BucketCreationJob(adapter.Resource())
	if err := r.createKubernetesResource(ctx, buckets, adapter); err != nil {
		return err
	}

	// migration := gitlabctl.MigrationsJob(cr)
	// return r.createKubernetesResource(migration, cr)

	return r.runMigrationsJob(ctx, adapter)
}

func (r *GitLabReconciler) reconcileServiceMonitor(ctx context.Context, adapter helpers.CustomResourceAdapter) error {
	var servicemonitors []*monitoringv1.ServiceMonitor

	gitaly := gitlabctl.GitalyServiceMonitor(adapter.Resource())

	gitlab := gitlabctl.ExporterServiceMonitor(adapter.Resource())

	postgres := gitlabctl.PostgresqlServiceMonitor(adapter.Resource())

	redis := gitlabctl.RedisServiceMonitor(adapter.Resource())

	workhorse := gitlabctl.WebserviceServiceMonitor(adapter.Resource())

	servicemonitors = append(servicemonitors,
		gitlab,
		gitaly,
		postgres,
		redis,
		workhorse,
	)

	for _, sm := range servicemonitors {
		if err := r.createKubernetesResource(ctx, sm, adapter); err != nil {
			return err
		}
	}

	service := gitlabctl.ExposePrometheusCluster(adapter.Resource())
	if err := r.createKubernetesResource(ctx, service, nil); err != nil {
		return err
	}

	prometheus := gitlabctl.PrometheusCluster(adapter.Resource())
	return r.createKubernetesResource(ctx, prometheus, nil)
}

func (r *GitLabReconciler) runMigrationsJob(ctx context.Context, adapter helpers.CustomResourceAdapter) error {
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
	if err := r.createKubernetesResource(ctx, migrations, adapter); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcileDeployments(ctx context.Context, adapter helpers.CustomResourceAdapter) error {

	if err := r.reconcileWebserviceDeployment(ctx, adapter); err != nil {
		return err
	}

	if err := r.reconcileShellDeployment(ctx, adapter); err != nil {
		return err
	}

	if err := r.reconcileSidekiqDeployment(ctx, adapter); err != nil {
		return err
	}

	if err := r.reconcileRegistryDeployment(ctx, adapter); err != nil {
		return err
	}

	if err := r.reconcileTaskRunnerDeployment(ctx, adapter); err != nil {
		return err
	}

	if err := r.reconcileGitlabExporterDeployment(ctx, adapter); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcileStatefulSets(ctx context.Context, adapter helpers.CustomResourceAdapter) error {

	var statefulsets []*appsv1.StatefulSet

	gitaly := gitlabctl.GitalyStatefulSet(adapter)
	redis := gitlabctl.RedisStatefulSet(adapter)
	postgres := gitlabctl.PostgresStatefulSet(adapter)

	statefulsets = append(statefulsets, postgres, redis, gitaly)

	for _, statefulset := range statefulsets {
		if err := r.createKubernetesResource(ctx, statefulset, adapter); err != nil {
			return err
		}
	}

	return nil
}

func (r *GitLabReconciler) createKubernetesResource(ctx context.Context, object interface{}, adapter helpers.CustomResourceAdapter) error {

	if r.isObjectFound(object) {
		return nil
	}

	// If parent resource is nil, not owner reference will be set
	if adapter != nil && adapter.Resource() != nil {
		if err := controllerutil.SetControllerReference(adapter.Resource(), object.(metav1.Object), r.Scheme); err != nil {
			return err
		}
	}

	return r.Create(ctx, object.(runtime.Object).DeepCopyObject())
}

// TODO: Remove this function
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

func (r *GitLabReconciler) reconcileMinioInstance(ctx context.Context, adapter helpers.CustomResourceAdapter) error {
	cm := gitlabctl.MinioScriptConfigMap(adapter.Resource())
	if err := r.createKubernetesResource(ctx, cm, adapter); err != nil {
		return err
	}

	secret := gitlabctl.MinioSecret(adapter.Resource())
	if err := r.createKubernetesResource(ctx, secret, adapter); err != nil && errors.IsAlreadyExists(err) {
		return err
	}

	appConfigSecret, err := gitlabctl.AppConfigConnectionSecret(adapter, *secret)
	if err != nil {
		return err
	}

	if err := r.createKubernetesResource(ctx, appConfigSecret, adapter); err != nil && errors.IsAlreadyExists(err) {
		return err
	}

	registryConnectionSecret, err := gitlabctl.RegistryConnectionSecret(adapter, *secret)
	if err != nil {
		return err
	}

	if err := r.createKubernetesResource(ctx, registryConnectionSecret, adapter); err != nil && errors.IsAlreadyExists(err) {
		return err
	}

	taskRunnerConnectionSecret := gitlabctl.TaskRunnerConnectionSecret(adapter, *secret)
	if err := r.createKubernetesResource(ctx, taskRunnerConnectionSecret, adapter); err != nil && errors.IsAlreadyExists(err) {
		return err
	}

	// Only deploy the minio service and statefulset for development builds
	if adapter.Resource().Spec.ObjectStore.Development {
		svc := gitlabctl.MinioService(adapter.Resource())
		if err := r.createKubernetesResource(ctx, svc, adapter); err != nil {
			return err
		}

		// deploy minio
		minio := gitlabctl.MinioStatefulSet(adapter.Resource())
		return r.createKubernetesResource(ctx, minio, adapter)
	}

	return nil
}

func (r *GitLabReconciler) reconcileServices(ctx context.Context, adapter helpers.CustomResourceAdapter) error {
	var services []*corev1.Service

	shell := gitlabctl.ShellService(adapter)
	gitaly := gitlabctl.GitalyService(adapter)
	exporter := gitlabctl.ExporterService(adapter)
	webservice := gitlabctl.WebserviceService(adapter)
	redis := gitlabctl.RedisServices(adapter)
	postgres := gitlabctl.PostgresServices(adapter)
	registry := gitlabctl.RegistryService(adapter)

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
		if err := r.createKubernetesResource(ctx, svc, adapter); err != nil {
			return err
		}
	}

	return nil
}

func (r *GitLabReconciler) reconcileGitlabExporterDeployment(ctx context.Context, adapter helpers.CustomResourceAdapter) error {

	exporter := gitlabctl.ExporterDeployment(adapter)

	if err := controllerutil.SetControllerReference(adapter.Resource(), exporter, r.Scheme); err != nil {
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

func (r *GitLabReconciler) reconcileWebserviceDeployment(ctx context.Context, adapter helpers.CustomResourceAdapter) error {
	webservice := gitlabctl.WebserviceDeployment(adapter)

	if err := controllerutil.SetControllerReference(adapter.Resource(), webservice, r.Scheme); err != nil {
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

func (r *GitLabReconciler) reconcileRegistryDeployment(ctx context.Context, adapter helpers.CustomResourceAdapter) error {
	registry := gitlabctl.RegistryDeployment(adapter)

	if err := controllerutil.SetControllerReference(adapter.Resource(), registry, r.Scheme); err != nil {
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

func (r *GitLabReconciler) reconcileShellDeployment(ctx context.Context, adapter helpers.CustomResourceAdapter) error {
	shell := gitlabctl.ShellDeployment(adapter)

	if err := controllerutil.SetControllerReference(adapter.Resource(), shell, r.Scheme); err != nil {
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

func (r *GitLabReconciler) reconcileSidekiqDeployment(ctx context.Context, adapter helpers.CustomResourceAdapter) error {
	sidekiq := gitlabctl.SidekiqDeployment(adapter)

	if err := controllerutil.SetControllerReference(adapter.Resource(), sidekiq, r.Scheme); err != nil {
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

func (r *GitLabReconciler) reconcileTaskRunnerDeployment(ctx context.Context, adapter helpers.CustomResourceAdapter) error {
	tasker := gitlabctl.TaskRunnerDeployment(adapter)

	if err := controllerutil.SetControllerReference(adapter.Resource(), tasker, r.Scheme); err != nil {
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

func (r *GitLabReconciler) exposeGitLabInstance(ctx context.Context, adapter helpers.CustomResourceAdapter) error {
	// if gitlabutils.IsOpenshift() {
	// 	return r.reconcileRoute(cr)
	// }

	return r.reconcileIngress(ctx, adapter)
}

func (r *GitLabReconciler) reconcileRoute(ctx context.Context, adapter helpers.CustomResourceAdapter) error {
	app := gitlabctl.MainRoute(adapter.Resource())

	admin := gitlabctl.AdminRoute(adapter.Resource())

	registry := gitlabctl.RegistryRoute(adapter.Resource())

	var routes []*routev1.Route
	routes = append(routes,
		app,
		admin,
		registry,
	)

	for _, route := range routes {
		if err := r.createKubernetesResource(ctx, route, adapter); err != nil {
			return err
		}
	}

	return nil
}

func (r *GitLabReconciler) reconcileIngress(ctx context.Context, adapter helpers.CustomResourceAdapter) error {

	controller := gitlabctl.IngressController(adapter.Resource())
	if err := r.createKubernetesResource(ctx, controller, adapter); err != nil {
		return err
	}

	var ingresses []*extensionsv1beta1.Ingress
	gitlab := gitlabctl.Ingress(adapter.Resource())
	registry := gitlabctl.RegistryIngress(adapter.Resource())

	ingresses = append(ingresses,
		gitlab,
		registry,
	)

	for _, ingress := range ingresses {
		if err := r.createKubernetesResource(ctx, ingress, adapter); err != nil {
			return err
		}
	}

	return nil
}

func (r *GitLabReconciler) reconcileCertManagerCertificates(ctx context.Context, adapter helpers.CustomResourceAdapter) error {
	// certificates := RequiresCertificate(cr)

	issuer := gitlabctl.CertificateIssuer(adapter.Resource())

	return r.createKubernetesResource(ctx, issuer, adapter)
}

func (r *GitLabReconciler) setupAutoscaling(ctx context.Context, adapter helpers.CustomResourceAdapter) error {
	selector := client.MatchingLabelsSelector{
		Selector: getLabelSet(adapter.Resource()).AsSelector(),
	}

	deployments := &appsv1.DeploymentList{}
	err := r.List(ctx, deployments, client.InNamespace(adapter.Resource().Namespace), selector)
	if err != nil {
		return err
	}

	for _, deploy := range deployments.Items {
		if err := r.reconcileHPA(ctx, &deploy, adapter); err != nil {
			return err
		}
	}

	return nil
}

func (r *GitLabReconciler) reconcileHPA(ctx context.Context, deployment *appsv1.Deployment, adapter helpers.CustomResourceAdapter) error {
	excludedDeployments := [2]string{"gitlab-exporter", "gitlab-task-runner"}
	for _, excludedDeployment := range excludedDeployments {
		if strings.Contains(deployment.Name, excludedDeployment) {
			return nil
		}
	}

	hpa := gitlabctl.HorizontalAutoscaler(deployment, adapter.Resource())

	found := &autoscalingv1.HorizontalPodAutoscaler{}
	err := r.Get(ctx, types.NamespacedName{Name: deployment.Name, Namespace: adapter.Resource().Namespace}, found)
	if err != nil {
		if errors.IsNotFound(err) {
			// return nil if hpa is nil
			if hpa == nil {
				return nil
			}

			if err := controllerutil.SetControllerReference(adapter.Resource(), hpa, r.Scheme); err != nil {
				return err
			}

			return r.Create(ctx, hpa.DeepCopy())
		}

		return err
	}

	if adapter.Resource().Spec.AutoScaling == nil {
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

func (r *GitLabReconciler) reconcileServiceAccount(ctx context.Context, adapter helpers.CustomResourceAdapter) error {
	sa := gitlabutils.ServiceAccount("gitlab-app", adapter.Namespace())

	found := &corev1.ServiceAccount{}
	lookupKey := types.NamespacedName{Name: sa.Name, Namespace: adapter.Namespace()}
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
