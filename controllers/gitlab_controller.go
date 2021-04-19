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
	"time"

	"github.com/go-logr/logr"
	"github.com/imdario/mergo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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

	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
	gitlabctl "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/gitlab"
	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/internal"

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
// +kubebuilder:rbac:groups=core,resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
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

// Reconcile triggers when an event occurs on the watched resource
func (r *GitLabReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("gitlab", req.NamespacedName)

	log.Info("Reconciling GitLab")
	gitlab := &gitlabv1beta1.GitLab{}
	if err := r.Get(ctx, req.NamespacedName, gitlab); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		// could not get GitLab resource
		return ctrl.Result{}, err
	}

	adapter := gitlabctl.NewCustomResourceAdapter(gitlab)

	if err := r.reconcileServiceAccount(ctx, adapter); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileNamespaces(ctx); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.runSharedSecretsJob(ctx, adapter); err != nil {
		return ctrl.Result{}, err
	}

	tlsSecretName, err := gitlabctl.GetStringValue(adapter.Values(), "global.ingress.tls.secretName")
	if err != nil || tlsSecretName == "" {
		if err := r.runSelfSignedCertsJob(ctx, adapter); err != nil {
			return ctrl.Result{}, err
		}
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

	if internal.RequiresCertManagerCertificate(adapter).Any() {
		if err := r.reconcileCertManagerCertificates(ctx, adapter); err != nil {
			return ctrl.Result{}, err
		}
	}

	waitInterval := 5 * time.Second
	if !r.ifCoreServicesReady(ctx, adapter) {
		log.Info("Core services are not ready. Waiting and retrying", "interval", waitInterval)
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

	if internal.IsPrometheusSupported() {
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
		Complete(r)
}

func (r *GitLabReconciler) runSharedSecretsJob(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	cfgMap, job, err := gitlabctl.SharedSecretsResources(adapter)
	if err != nil {
		return err
	}

	if _, err := r.createOrPatch(ctx, cfgMap, adapter); err != nil {
		return err
	}

	return r.runJobAndWait(ctx, adapter, job)
}

func (r *GitLabReconciler) runSelfSignedCertsJob(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	job, err := gitlabctl.SelfSignedCertsJob(adapter)
	if err != nil {
		return err
	}

	return r.runJobAndWait(ctx, adapter, job)
}

func (r *GitLabReconciler) runJobAndWait(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, job *batchv1.Job) error {

	logger := r.Log.WithValues("gitlab", adapter.Reference(), "job", job.Name, "namespace", job.Namespace)

	_, err := r.createOrPatch(ctx, job, adapter)
	if err != nil {
		return err
	}

	elapsed := time.Duration(0)
	timeout := gitlabctl.SharedSecretsJobTimeout()
	waitPeriod := gitlabctl.SharedSecretsJobWaitPeriod(timeout, elapsed)
	lookupKey := types.NamespacedName{
		Name:      job.Name,
		Namespace: job.Namespace,
	}

	var result error = nil

	for {
		if elapsed > timeout {
			result = errors.NewTimeoutError("The Job did not finish in time", int(timeout))
			logger.Error(result, "Timeout for Job exceeded.",
				"timeout", timeout)
			break
		}

		logger.V(2).Info("Checking the status of Job")
		lookupVal := &batchv1.Job{}
		if err := r.Get(context.Background(), lookupKey, lookupVal); err != nil {
			logger.V(2).Info("Failed to check the status of Job", "error", err)

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
			logger.V(2).Info("Job succeeded")
			break
		}

		if lookupVal.Status.Failed > 0 {
			result = errors.NewInternalError(
				fmt.Errorf("job %s has failed, check the logs in %s", job.Name, lookupKey))
			logger.Error(result, "Job failed")
			break
		}

		elapsed += waitPeriod
		time.Sleep(waitPeriod)
	}

	return result
}

//	Reconciler for all ConfigMaps come below
func (r *GitLabReconciler) reconcileConfigMaps(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
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

	configmaps = append(configmaps,
		gitaly,
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
		if _, err := r.createOrPatch(ctx, cm, adapter); err != nil {
			return err
		}
	}

	return nil
}

func (r *GitLabReconciler) reconcileJobs(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {

	// initialize buckets once s3 storage is up
	buckets := internal.BucketCreationJob(adapter)
	if _, err := r.createOrPatch(ctx, buckets, adapter); err != nil {
		return err
	}

	// migration := gitlabctl.MigrationsJob(cr)
	// return r.createOrPatch(migration, cr)

	return r.runMigrationsJob(ctx, adapter)
}

func (r *GitLabReconciler) reconcileServiceMonitor(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	var servicemonitors []*monitoringv1.ServiceMonitor

	gitaly := internal.GitalyServiceMonitor(adapter.Resource())

	gitlab := internal.ExporterServiceMonitor(adapter.Resource())

	postgres := internal.PostgresqlServiceMonitor(adapter.Resource())

	redis := internal.RedisServiceMonitor(adapter.Resource())

	workhorse := internal.WebserviceServiceMonitor(adapter.Resource())

	servicemonitors = append(servicemonitors,
		gitlab,
		gitaly,
		postgres,
		redis,
		workhorse,
	)

	for _, sm := range servicemonitors {
		if _, err := r.createOrPatch(ctx, sm, adapter); err != nil {
			return err
		}
	}

	service := internal.ExposePrometheusCluster(adapter.Resource())
	if _, err := r.createOrPatch(ctx, service, adapter); err != nil {
		return err
	}

	prometheus := internal.PrometheusCluster(adapter.Resource())

	_, err := r.createOrPatch(ctx, prometheus, adapter)
	return err
}

func (r *GitLabReconciler) runMigrationsJob(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	migrations, err := gitlabctl.MigrationsJob(adapter)
	if err != nil {
		return err
	}

	if _, err := r.createOrPatch(ctx, migrations, adapter); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcileDeployments(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {

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

func (r *GitLabReconciler) reconcileStatefulSets(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {

	var statefulsets []*appsv1.StatefulSet

	gitaly := gitlabctl.GitalyStatefulSet(adapter)
	redis := gitlabctl.RedisStatefulSet(adapter)
	postgres := gitlabctl.PostgresStatefulSet(adapter)

	statefulsets = append(statefulsets, postgres, redis, gitaly)

	for _, statefulset := range statefulsets {
		if _, err := r.createOrPatch(ctx, statefulset, adapter); err != nil {
			return err
		}
	}

	return nil
}

var ignoreObjectMetaFields = []string{
	"generateName",
	"finalizers",
	"clusterName",
	"managedFields",

	"uid",
	"resourceVersion",
	"generation",
	"creationTimestamp",
	"deletionTimestamp",
	"deletionGracePeriodSeconds",
	"clusterName",
}

func mutateObject(source, target client.Object) (err error) {
	sourceFullName, targetFullName :=
		fmt.Sprintf("%s/%s", source.GetName(), source.GetNamespace()),
		fmt.Sprintf("%s/%s", target.GetName(), target.GetNamespace())
	if sourceFullName != targetFullName {
		err = fmt.Errorf("source and target must refer to the same object: %s, %s",
			sourceFullName, targetFullName)
		return
	}

	// Map both source and target to Unstructured for further untyped manipulation.
	src, err := runtime.DefaultUnstructuredConverter.ToUnstructured(source)
	if err != nil {
		return
	}

	dst, err := runtime.DefaultUnstructuredConverter.ToUnstructured(target)
	if err != nil {
		return
	}

	// Remove status from source to make sure that
	// the source does not have any immutable metadata field.
	unstructured.RemoveNestedField(src, "status")

	for _, f := range ignoreObjectMetaFields {
		unstructured.RemoveNestedField(src, "metadata", f)
	}

	// TODO: Handle other immutable attributes, e.g. .sepc.selector

	// Merge source into target.
	if err = mergo.Merge(&dst, src, mergo.WithOverride); err != nil {
		return
	}

	// Map the target back to type object.
	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(dst, target); err != nil {
		return
	}

	return
}

func (r *GitLabReconciler) createOrPatch(ctx context.Context, templateObject client.Object, adapter gitlabctl.CustomResourceAdapter) (applied bool, err error) {
	applied = false
	err = nil

	if templateObject == nil {
		r.Log.Info("Controller is not able to delete managed resources. This is a known issue",
			"gitlab", adapter.Reference())
	}

	key := client.ObjectKeyFromObject(templateObject)

	logger := r.Log.WithValues(
		"gitlab", adapter.Reference(),
		"type", fmt.Sprintf("%T", templateObject),
		"reference", key)

	logger.V(2).Info("Setting controller reference")
	if err = controllerutil.SetControllerReference(adapter.Resource(), templateObject, r.Scheme); err != nil {
		return
	}

	existing := templateObject.DeepCopyObject().(client.Object)

	if err = r.Get(ctx, key, existing); err != nil {
		if !errors.IsNotFound(err) {
			return
		}

		logger.V(1).Info("Creating object")
		err = r.Create(ctx, existing)
		applied = err == nil
		return
	}

	// If Secret and related to MinIO, skip the patch.
	// TODO: replace MinIO generated secrets (along with other objects) with MinIO chart or Operator.
	if existing.GetLabels()["app.kubernetes.io/component"] == "minio" && existing.GetObjectKind().GroupVersionKind().Kind == "Secret" {
		return
	}

	mutate := func() error {
		return mutateObject(templateObject, existing)
	}

	result, err := controllerutil.CreateOrPatch(ctx, r.Client, existing, mutate)
	if err != nil {
		return
	}

	applied = true
	logger.V(1).Info("createOrPatch result", "result", result)

	return
}

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

	taskRunnerConnectionSecret := internal.TaskRunnerConnectionSecret(adapter, *secret)
	if _, err := r.createOrPatch(ctx, taskRunnerConnectionSecret, adapter); err != nil && errors.IsAlreadyExists(err) {
		return err
	}

	// Only deploy the minio service and statefulset for development builds
	if minioEnabled, _ := gitlabctl.GetBoolValue(adapter.Values(), "global.appConfig.object_store.enabled"); minioEnabled {
		ing := internal.MinioIngress(adapter)
		if _, err := r.createOrPatch(ctx, ing, adapter); err != nil {
			return err
		}

		svc := internal.MinioService(adapter)
		if _, err := r.createOrPatch(ctx, svc, adapter); err != nil {
			return err
		}

		// deploy minio
		minio := internal.MinioStatefulSet(adapter)
		_, err := r.createOrPatch(ctx, minio, adapter)
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcileServices(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
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
		if _, err := r.createOrPatch(ctx, svc, adapter); err != nil {
			return err
		}
	}

	return nil
}

func (r *GitLabReconciler) reconcileGitlabExporterDeployment(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	exporter := gitlabctl.ExporterDeployment(adapter)

	_, err := r.createOrPatch(ctx, exporter, adapter)

	return err
}

func (r *GitLabReconciler) reconcileWebserviceDeployment(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	webservice := gitlabctl.WebserviceDeployment(adapter)

	_, err := r.createOrPatch(ctx, webservice, adapter)

	return err
}

func (r *GitLabReconciler) reconcileRegistryDeployment(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	registry := gitlabctl.RegistryDeployment(adapter)

	_, err := r.createOrPatch(ctx, registry, adapter)

	return err
}

func (r *GitLabReconciler) reconcileShellDeployment(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	shell := gitlabctl.ShellDeployment(adapter)

	_, err := r.createOrPatch(ctx, shell, adapter)

	return err
}

func (r *GitLabReconciler) reconcileSidekiqDeployment(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	sidekiq := gitlabctl.SidekiqDeployment(adapter)

	_, err := r.createOrPatch(ctx, sidekiq, adapter)

	return err
}

func (r *GitLabReconciler) reconcileTaskRunnerDeployment(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	tasker := gitlabctl.TaskRunnerDeployment(adapter)

	_, err := r.createOrPatch(ctx, tasker, adapter)

	return err
}

func (r *GitLabReconciler) exposeGitLabInstance(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	// if internal.IsOpenshift() {
	// 	return r.reconcileRoute(cr)
	// }

	return r.reconcileIngress(ctx, adapter)
}

func (r *GitLabReconciler) reconcileIngress(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	var ingresses []*extensionsv1beta1.Ingress
	gitlab := gitlabctl.WebserviceIngress(adapter)
	registry := gitlabctl.RegistryIngress(adapter)

	ingresses = append(ingresses,
		gitlab,
		registry,
	)

	for _, ingress := range ingresses {
		if _, err := r.createOrPatch(ctx, ingress, adapter); err != nil {
			return err
		}
	}

	return nil
}

func (r *GitLabReconciler) reconcileCertManagerCertificates(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	// certificates := RequiresCertificate(cr)

	issuer := internal.CertificateIssuer(adapter)

	_, err := r.createOrPatch(ctx, issuer, adapter)
	return err
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

func (r *GitLabReconciler) reconcileServiceAccount(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	sa := internal.ServiceAccount("gitlab-app", adapter.Namespace())

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

func (r *GitLabReconciler) isEndpointReady(ctx context.Context, service string, adapter gitlabctl.CustomResourceAdapter) bool {
	var addresses []corev1.EndpointAddress

	ep := &corev1.Endpoints{}
	err := r.Get(ctx, types.NamespacedName{Name: service, Namespace: adapter.Namespace()}, ep)
	if err != nil && errors.IsNotFound(err) {
		return false
	}

	for _, subset := range ep.Subsets {
		addresses = append(addresses, subset.Addresses...)
	}

	return len(addresses) > 0
}

func (r *GitLabReconciler) ifCoreServicesReady(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) bool {
	return r.isEndpointReady(ctx, adapter.ReleaseName()+"-postgresql", adapter) &&
		r.isEndpointReady(ctx, adapter.ReleaseName()+"-gitaly", adapter) &&
		r.isEndpointReady(ctx, adapter.ReleaseName()+"-redis-master", adapter)
}
