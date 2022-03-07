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
	"regexp"
	"time"

	"github.com/go-logr/logr"
	"github.com/imdario/mergo"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"

	certmanagerv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"

	gitlabv1beta1 "gitlab.com/gitlab-org/cloud-native/gitlab-operator/api/v1beta1"
	gitlabctl "gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/gitlab"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/internal"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/settings"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"

	"k8s.io/apimachinery/pkg/api/errors"
)

// GitLabReconciler reconciles a GitLab object.
type GitLabReconciler struct {
	client.Client

	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
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
// +kubebuilder:rbac:groups=batch,resources=cronjobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=servicemonitors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=prometheuses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cert-manager.io,resources=issuers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cert-manager.io,resources=certificates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=route.openshift.io,resources=routes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=route.openshift.io,resources=routes/custom-host,verbs=get;list;watch;create;update;patch;delete

// Reconcile triggers when an event occurs on the watched resource.
//nolint:gocognit,gocyclo,nestif // The complexity of this method will be addressed in #260.
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

	isUpgrade := adapter.IsUpgrade()
	log.V(1).Info("version information", "upgrade", isUpgrade, "current version", adapter.StatusVersion(), "desired version", adapter.ChartVersion())

	if err := r.setStatusCondition(ctx, adapter, ConditionInitialized, false, "GitLab is initializing"); err != nil {
		return ctrl.Result{}, err
	}

	template, err := gitlabctl.GetTemplate(adapter)
	if err != nil {
		r.Recorder.Event(adapter.Resource(), "Warning", "ConfigError",
			fmt.Sprintf("Configuration error detected: %v", err))
		return ctrl.Result{}, nil // return nil here to prevent further reconcile loops
	}

	if err := r.setStatusCondition(ctx, adapter, ConditionInitialized, true, "GitLab is initialized"); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.setStatusCondition(ctx, adapter, ConditionAvailable, false, "GitLab is starting but not yet available"); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileServiceAccount(ctx, adapter); err != nil {
		return ctrl.Result{}, err
	}

	if gitlabctl.NGINXEnabled(adapter) {
		if err := r.reconcileNGINX(ctx, adapter, template); err != nil {
			return ctrl.Result{}, err
		}
	}

	if err := r.runSharedSecretsJob(ctx, adapter, template); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.runSelfSignedCertsJob(ctx, adapter, template); err != nil {
		return ctrl.Result{}, err
	}

	if gitlabctl.PostgresEnabled(adapter) {
		if err := r.reconcilePostgres(ctx, adapter, template); err != nil {
			return ctrl.Result{}, err
		}
	} else {
		if err := r.validateExternalPostgresConfiguration(ctx, adapter); err != nil {
			return ctrl.Result{}, err
		}
	}

	if gitlabctl.RedisEnabled(adapter) {
		if err := r.reconcileRedis(ctx, adapter, template); err != nil {
			return ctrl.Result{}, err
		}
	} else {
		if err := r.validateExternalRedisConfiguration(ctx, adapter); err != nil {
			return ctrl.Result{}, err
		}
	}

	if gitlabctl.GitalyEnabled(adapter) {
		if !gitlabctl.PraefectEnabled(adapter) || !gitlabctl.PraefectReplaceInternalGitalyEnabled(adapter) {
			if err := r.reconcileGitaly(ctx, adapter, template); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	if gitlabctl.PraefectEnabled(adapter) {
		if err := r.reconcilePraefect(ctx, adapter, template); err != nil {
			return ctrl.Result{}, err
		}

		if gitlabctl.GitalyEnabled(adapter) {
			if err := r.reconcileGitalyPraefect(ctx, adapter, template); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	if gitlabctl.MinioEnabled(adapter) {
		if err := r.reconcileMinioInstance(ctx, adapter); err != nil {
			return ctrl.Result{}, err
		}
	}

	if gitlabctl.MailroomEnabled(adapter) {
		if err := r.reconcileMailroom(ctx, adapter, template); err != nil {
			return ctrl.Result{}, err
		}
	}

	if internal.RequiresCertManagerCertificate(adapter).Any() {
		if err := r.reconcileCertManagerCertificates(ctx, adapter); err != nil {
			return ctrl.Result{}, err
		}
	}

	waitInterval := 5 * time.Second
	if !r.ifCoreServicesReady(ctx, adapter, template) {
		log.Info("Core services are not ready. Waiting and retrying", "interval", waitInterval)
		return ctrl.Result{RequeueAfter: waitInterval}, nil
	}

	if gitlabctl.ShellEnabled(adapter) {
		if err := r.reconcileGitLabShell(ctx, adapter, template); err != nil {
			return ctrl.Result{}, err
		}
	}

	if gitlabctl.RegistryEnabled(adapter) {
		if err := r.reconcileRegistry(ctx, adapter, template); err != nil {
			return ctrl.Result{}, err
		}
	}

	if gitlabctl.ToolboxEnabled(adapter) {
		if err := r.reconcileToolbox(ctx, adapter, template); err != nil {
			return ctrl.Result{}, err
		}
	}

	if gitlabctl.ExporterEnabled(adapter) {
		if err := r.reconcileGitLabExporter(ctx, adapter, template); err != nil {
			return ctrl.Result{}, err
		}
	}

	if gitlabctl.PagesEnabled(adapter) {
		if err := r.reconcilePages(ctx, adapter, template); err != nil {
			return ctrl.Result{}, err
		}
	}

	if gitlabctl.KasEnabled(adapter) {
		if err := r.reconcileKas(ctx, adapter, template); err != nil {
			return ctrl.Result{}, err
		}
	}

	if gitlabctl.MigrationsEnabled(adapter) {
		if err := r.reconcileMigrationsConfigMap(ctx, adapter, template); err != nil {
			return ctrl.Result{}, err
		}
	}

	if gitlabctl.SidekiqEnabled(adapter) {
		if err := r.reconcileSidekiqConfigMaps(ctx, adapter, template); err != nil {
			return ctrl.Result{}, err
		}
	}

	if gitlabctl.WebserviceEnabled(adapter) {
		if err := r.reconcileWebserviceExceptDeployments(ctx, adapter, template); err != nil {
			return ctrl.Result{}, err
		}
	}

	if isUpgrade {
		if err := r.setStatusCondition(ctx, adapter, ConditionUpgrading, true, fmt.Sprintf("GitLab is upgrading from %s to %s", adapter.StatusVersion(), adapter.ChartVersion())); err != nil {
			return ctrl.Result{}, err
		}

		if gitlabctl.MigrationsEnabled(adapter) {
			if gitlabctl.WebserviceEnabled(adapter) || gitlabctl.SidekiqEnabled(adapter) {
				// If upgrading with Migrations enabled and Webservice and/or Sidekiq enabled,
				// then follow the traditional upgrade logic.
				if err := r.reconcileWebserviceAndSidekiqIfEnabled(ctx, adapter, template, true, log); err != nil {
					return ctrl.Result{}, err
				}

				log.Info("reconciling pre migrations")

				if err := r.runPreMigrations(ctx, adapter, template); err != nil {
					return ctrl.Result{}, err
				}

				if err := r.unpauseWebserviceAndSidekiqIfEnabled(ctx, adapter, template, log); err != nil {
					return ctrl.Result{}, err
				}

				if err := r.webserviceAndSidekiqRunningIfEnabled(ctx, adapter, template, log); err != nil {
					return ctrl.Result{}, err
				}

				log.Info("reconciling post migrations")

				if err := r.runAllMigrations(ctx, adapter, template); err != nil {
					return ctrl.Result{}, err
				}

				if err := r.rollingUpdateWebserviceAndSidekiqIfEnabled(ctx, adapter, template, log); err != nil {
					return ctrl.Result{}, err
				}
			} else {
				// If upgrading with Migrations enabled but neither Webservice nor Sidekiq are enabled,
				// then just run all migrations.
				log.Info("running all migrations")

				if err := r.runAllMigrations(ctx, adapter, template); err != nil {
					return ctrl.Result{}, err
				}
			}
		} else {
			// If upgrading with Migrations disabled, then just reconcile enabled Deployments.
			if err := r.reconcileWebserviceAndSidekiqIfEnabled(ctx, adapter, template, false, log); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// If not upgrading, then run all migrations (if enabled) and reconcile enabled Deployments.
		if err := r.setStatusCondition(ctx, adapter, ConditionUpgrading, false, "GitLab is not currently upgrading"); err != nil {
			return ctrl.Result{}, err
		}

		if gitlabctl.MigrationsEnabled(adapter) {
			log.Info("running all migrations")

			if err := r.runAllMigrations(ctx, adapter, template); err != nil {
				return ctrl.Result{}, err
			}
		}

		if err := r.reconcileWebserviceAndSidekiqIfEnabled(ctx, adapter, template, false, log); err != nil {
			return ctrl.Result{}, err
		}
	}

	if err := r.setupAutoscaling(ctx, adapter, template); err != nil {
		return ctrl.Result{}, err
	}

	if settings.IsGroupVersionSupported("monitoring.coreos.com", "v1") {
		// Deploy a prometheus service monitor
		if err := r.reconcileServiceMonitor(ctx, adapter); err != nil {
			return ctrl.Result{}, err
		}
	}

	result, err := r.reconcileGitLabStatus(ctx, adapter, template)

	return result, err
}

// SetupWithManager configures the custom resource watched resources.
func (r *GitLabReconciler) SetupWithManager(mgr ctrl.Manager) error {
	builder := ctrl.NewControllerManagedBy(mgr).
		For(&gitlabv1beta1.GitLab{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&appsv1.Deployment{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&batchv1.Job{}).
		Owns(&batchv1beta1.CronJob{}).
		Owns(&networkingv1.Ingress{}).
		WithEventFilter(predicate.GenerationChangedPredicate{})

	if settings.IsGroupVersionSupported("monitoring.coreos.com", "v1") {
		r.Log.Info("Using monitoring.coreos.com/v1")
		builder.Owns(&monitoringv1.ServiceMonitor{})
	}

	if settings.IsGroupVersionSupported("cert-manager.io", "v1") {
		r.Log.Info("Using cert-manager.io/v1")
		builder.
			Owns(&certmanagerv1.Issuer{}).
			Owns(&certmanagerv1.Certificate{})
	}

	return builder.Complete(r)
}

func (r *GitLabReconciler) runJobAndWait(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, job client.Object, timeout time.Duration) error {
	logger := r.Log.WithValues("gitlab", adapter.Reference(), "job", job.GetName(), "namespace", job.GetNamespace())

	_, err := r.createOrPatch(ctx, job, adapter)
	if err != nil {
		return err
	}

	elapsed := time.Duration(0)
	waitPeriod := timeout / 100
	lookupKey := types.NamespacedName{
		Name:      job.GetName(),
		Namespace: job.GetNamespace(),
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
				fmt.Errorf("job %s has failed, check the logs in %s", job.GetName(), lookupKey))
			logger.Error(result, "Job failed")

			break
		}

		elapsed += waitPeriod
		time.Sleep(waitPeriod)
	}

	return result
}

func (r *GitLabReconciler) reconcileServiceMonitor(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	var servicemonitors []*monitoringv1.ServiceMonitor

	gitaly := internal.GitalyServiceMonitor(adapter.Resource())

	gitlab := internal.ExporterServiceMonitor(adapter.Resource())

	workhorse := internal.WebserviceServiceMonitor(adapter.Resource())

	servicemonitors = append(servicemonitors,
		gitlab,
		gitaly,
		workhorse,
	)

	if gitlabctl.RedisEnabled(adapter) {
		redis := internal.RedisServiceMonitor(adapter.Resource())
		servicemonitors = append(servicemonitors, redis)
	}

	if gitlabctl.PostgresEnabled(adapter) {
		postgres := internal.PostgresqlServiceMonitor(adapter.Resource())
		servicemonitors = append(servicemonitors, postgres)
	}

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

func mutateObject(source, target client.Object) error {
	sourceFullName, targetFullName :=
		fmt.Sprintf("%s/%s", source.GetName(), source.GetNamespace()),
		fmt.Sprintf("%s/%s", target.GetName(), target.GetNamespace())
	if sourceFullName != targetFullName {
		return fmt.Errorf("source and target must refer to the same object: %s, %s",
			sourceFullName, targetFullName)
	}

	// Map both source and target to Unstructured for further untyped manipulation.
	src, err := runtime.DefaultUnstructuredConverter.ToUnstructured(source)
	if err != nil {
		return err
	}

	dst, err := runtime.DefaultUnstructuredConverter.ToUnstructured(target)
	if err != nil {
		return err
	}

	// Remove status from source to make sure that
	// the source does not have any immutable metadata field.
	unstructured.RemoveNestedField(src, "status")

	for _, f := range ignoreObjectMetaFields {
		unstructured.RemoveNestedField(src, "metadata", f)
	}

	// Merge source into target.
	if err = mergo.Merge(&dst, src, mergo.WithOverride, mergo.WithSliceDeepCopy); err != nil {
		return err
	}

	// Map the target back to type object.
	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(dst, target); err != nil {
		return err
	}

	return nil
}

//nolint:unparam // The boolean return parameter is unused at the moment, but may be useful in the future.
func (r *GitLabReconciler) createOrPatch(ctx context.Context, templateObject client.Object, adapter gitlabctl.CustomResourceAdapter) (bool, error) {
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

	if err := controllerutil.SetControllerReference(adapter.Resource(), templateObject, r.Scheme); err != nil {
		return false, err
	}

	existing := templateObject.DeepCopyObject().(client.Object)

	if err := r.Get(ctx, key, existing); err != nil {
		if !errors.IsNotFound(err) {
			return false, err
		}

		logger.V(1).Info("Creating object")

		if err := r.Create(ctx, existing); err != nil {
			return false, err
		}

		return true, nil
	}

	// If Secret and related to MinIO, skip the patch.
	if existing.GetObjectKind().GroupVersionKind().Kind == "Secret" && existing.GetLabels()["app.kubernetes.io/component"] == "minio" && gitlabctl.MinioEnabled(adapter) {
		return false, nil
	}

	mutate := func() error {
		return mutateObject(templateObject, existing)
	}

	result, err := controllerutil.CreateOrPatch(ctx, r.Client, existing, mutate)
	if err != nil {
		return false, err
	}

	if result != controllerutil.OperationResultNone {
		logger.V(1).Info("createOrPatch result", "result", result)
	}

	return true, nil
}

func (r *GitLabReconciler) createOrUpdate(ctx context.Context, templateObject client.Object, adapter gitlabctl.CustomResourceAdapter) (bool, bool, error) {
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

	if err := controllerutil.SetControllerReference(adapter.Resource(), templateObject, r.Scheme); err != nil {
		return false, false, err
	}

	existing := templateObject.DeepCopyObject().(client.Object)

	if err := r.Get(ctx, key, existing); err != nil {
		if !errors.IsNotFound(err) {
			return false, false, err
		}

		if err := r.Create(ctx, existing); err != nil {
			return false, false, err
		}

		logger.V(1).Info("createOrUpdate result", "result", "created")

		return true, false, nil
	}

	templateObject.SetResourceVersion(existing.GetResourceVersion())

	if err := r.Update(ctx, templateObject); err != nil {
		logger.Error(err, "unable to update object", "object", templateObject)
		return false, false, err
	}

	logger.V(1).Info("createOrUpdate result", "result", "updated")

	return false, true, nil
}

func (r *GitLabReconciler) reconcileIngress(ctx context.Context, obj client.Object, adapter gitlabctl.CustomResourceAdapter) error {
	ingress, err := internal.AsIngress(obj)
	if err != nil {
		return err
	}

	logger := r.Log.WithValues("gitlab", adapter.Reference(), "namespace", adapter.Namespace())

	found := &networkingv1.Ingress{}
	err = r.Get(ctx, types.NamespacedName{Name: ingress.Name, Namespace: adapter.Namespace()}, found)

	if err != nil {
		if errors.IsNotFound(err) {
			logger.V(1).Info("creating ingress", "ingress", ingress.Name)
			return r.Create(ctx, ingress)
		}

		return err
	}

	// If resource is an Ingress and has an ACME challenge path, skip the patch.
	// This ensures that CertManager can add a path to existing ingresses for the ACME challenge without
	// the Operator immediately removing it before the challenge can be completed.
	doPatch := true
	regex := regexp.MustCompile("/.well-known/acme-challenge/+")

	for _, path := range found.Spec.Rules[0].IngressRuleValue.HTTP.Paths {
		if regex.MatchString(path.Path) {
			logger.V(1).Info("ingress contains ACME challenge path, skipping patch for now", "ingress", found.Name)

			doPatch = false
		}
	}

	if doPatch {
		if _, err := r.createOrPatch(ctx, ingress, adapter); err != nil {
			return err
		}
	}

	return nil
}

func (r *GitLabReconciler) reconcileCertManagerCertificates(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	issuer := internal.CertificateIssuer(adapter)

	_, _, err := r.createOrUpdate(ctx, issuer, adapter)

	return err
}

func (r *GitLabReconciler) reconcileServiceAccount(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	sa := internal.ServiceAccount("gitlab-app", adapter.Namespace())
	found := &corev1.ServiceAccount{}
	lookupKey := types.NamespacedName{Name: sa.Name, Namespace: adapter.Namespace()}

	if err := r.Get(ctx, lookupKey, found); err != nil {
		// gitlab-app ServiceAccount not found
		if errors.IsNotFound(err) {
			if err := r.Create(ctx, sa); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

func (r *GitLabReconciler) setupAutoscaling(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, template helm.Template) error {
	for _, hpa := range template.Query().ObjectsByKind(gitlabctl.HorizontalPodAutoscalerKind) {
		if _, err := r.createOrPatch(ctx, hpa, adapter); err != nil {
			return err
		}
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

func (r *GitLabReconciler) ifCoreServicesReady(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, template helm.Template) bool {
	if gitlabctl.PostgresEnabled(adapter) {
		if !r.isEndpointReady(ctx, adapter.ReleaseName()+"-postgresql", adapter) {
			return false
		}
	}

	if gitlabctl.RedisEnabled(adapter) {
		if !r.isEndpointReady(ctx, adapter.ReleaseName()+"-redis-master", adapter) {
			return false
		}
	}

	if gitlabctl.GitalyEnabled(adapter) {
		if !gitlabctl.PraefectEnabled(adapter) || !gitlabctl.PraefectReplaceInternalGitalyEnabled(adapter) {
			if !r.isEndpointReady(ctx, gitlabctl.GitalyService(template).GetName(), adapter) {
				return false
			}
		}
	}

	if gitlabctl.PraefectEnabled(adapter) {
		if !r.isEndpointReady(ctx, gitlabctl.PraefectService(template).GetName(), adapter) {
			return false
		}

		if gitlabctl.GitalyEnabled(adapter) {
			for _, gitalyPraefectService := range gitlabctl.GitalyPraefectServices(template) {
				if !r.isEndpointReady(ctx, gitalyPraefectService.GetName(), adapter) {
					return false
				}
			}
		}
	}

	return true
}

// If a Deployment has an HPA attached to it consult its Status to set the replica count.
func (r *GitLabReconciler) setDeploymentReplica(ctx context.Context, obj client.Object) error {
	deployment, err := internal.AsDeployment(obj)
	if err != nil {
		return err
	}

	// Get the Deployment's HPA so we can check the desired number of replicas.
	// Finds the Deployment's HPA using the Deployment's name (since they are defined the same way in the Helm chart).
	hpa := &autoscalingv1.HorizontalPodAutoscaler{}
	if err := r.Get(ctx, types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}, hpa); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}

		return err
	}

	replicas := hpa.Status.DesiredReplicas
	if replicas == 0 {
		return nil
	}

	if deployment.Spec.Replicas == nil || *(deployment.Spec.Replicas) != replicas {
		r.Log.V(1).Info("Changing replica count of deployment with HPA",
			"deployment", types.NamespacedName{
				Namespace: deployment.Namespace,
				Name:      deployment.Name,
			},
			"replicas", replicas)

		deployment.Spec.Replicas = &replicas
	}

	return nil
}

func (r *GitLabReconciler) annotateSecretsChecksum(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, obj client.Object) error {
	template, err := internal.GetPodTemplateSpec(obj)
	if err != nil {
		return err
	}

	secretsInfo := internal.PopulateAttachedSecrets(*template)
	for secretName, secretKeys := range secretsInfo {
		secret := &corev1.Secret{}
		lookupKey := types.NamespacedName{Name: secretName, Namespace: adapter.Namespace()}

		if err := r.Get(ctx, lookupKey, secret); err != nil {
			if errors.IsNotFound(err) {
				// Skip this Secret. Do not overreact to it being missing.
				continue
			}

			return err
		}

		hash := internal.SecretChecksum(*secret, secretKeys)
		if hash == "" {
			continue
		}

		if template.ObjectMeta.Annotations == nil {
			template.ObjectMeta.Annotations = map[string]string{}
		}

		template.ObjectMeta.Annotations[fmt.Sprintf("checksum/secret-%s", secretName)] = hash
	}

	return nil
}

func (r *GitLabReconciler) ensureSecret(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, secretName string) error {
	secret := &corev1.Secret{}
	lookupKey := types.NamespacedName{Name: secretName, Namespace: adapter.Namespace()}
	err := r.Get(ctx, lookupKey, secret)

	if err != nil {
		if errors.IsNotFound(err) {
			return fmt.Errorf("Secret '%s' not found", lookupKey)
		}

		return err
	}

	return nil
}
