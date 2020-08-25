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
	"sync"
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

	// nginxv1alpha1 "github.com/nginxinc/nginx-ingress-operator/pkg/apis/k8s/v1alpha1"
	certmanagerv1beta1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1beta1"
	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
	gitlabctl "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/gitlab"
	gitlabutils "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/utils"
	"k8s.io/apimachinery/pkg/api/errors"
)

// GitLabReconciler reconciles a GitLab object
type GitLabReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=apps.gitlab.com,namespace="placeholder",resources=gitlabs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.gitlab.com,namespace="placeholder",resources=gitlabs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,namespace="placeholder",resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,namespace="placeholder",resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,namespace="placeholder",resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,namespace="placeholder",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,namespace="placeholder",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,namespace="placeholder",resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,namespace="placeholder",resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,namespace="placeholder",resources=cronjobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=extensions,namespace="placeholder",resources=ingresses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=monitoring.coreos.com,namespace="placeholder",resources=servicemonitors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=monitoring.coreos.com,namespace="placeholder",resources=prometheuses,verbs=get;list;watch;create;update;patch;delete

// Reconcile triggers when an event occurs on the watched resource
func (r *GitLabReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("gitlab", req.NamespacedName)

	gitlab := &gitlabv1beta1.GitLab{}
	if err := r.Get(ctx, req.NamespacedName, gitlab); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		// could not get GitLab resource
		return ctrl.Result{}, err
	}

	if err := r.reconcileConfigMaps(gitlab); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileSecrets(gitlab); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileServices(gitlab); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.maskEmailPasword(gitlab); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileStatefulSets(gitlab); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileMinioInstance(gitlab); err != nil {
		return ctrl.Result{}, err
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		for !gitlabctl.IsEndpointReady(gitlab.Name+"-postgresql", gitlab) {
			time.Sleep(time.Second * 1)
		}
		wg.Done()
	}()

	// if RequiresCertManagerCertificate(gitlab).All() {
	// 	if err := r.reconcileCertManagerCertificates(gitlab); err != nil {
	// 		return err
	// 	}
	// }

	wg.Wait()

	if err := r.reconcileJobs(gitlab); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileDeployments(gitlab); err != nil {
		return ctrl.Result{}, err
	}

	// Deploy ingress to expose GitLab
	// if err := r.reconcileIngress(gitlab); err != nil {
	// 	return err
	// }

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
		Owns(&certmanagerv1beta1.Issuer{}).
		Owns(&certmanagerv1beta1.Certificate{}).
		// Owns(&nginxv1alpha1.NginxIngressController{}).
		Complete(r)
}

//	Reconciler for all ConfigMaps come below
func (r *GitLabReconciler) reconcileConfigMaps(cr *gitlabv1beta1.GitLab) error {
	var configmaps []*corev1.ConfigMap

	shell := gitlabctl.ShellConfigMap(cr)

	gitaly := gitlabctl.GitalyConfigMap(cr)

	redis := gitlabctl.RedisConfigMap(cr)

	redisScripts := gitlabctl.RedisSciptsConfigMap(cr)

	webservice := gitlabctl.WebserviceConfigMap(cr)

	workhorse := gitlabctl.WorkhorseConfigMap(cr)

	gitlab := gitlabctl.GetGitLabConfigMap(cr)

	sidekiq := gitlabctl.SidekiqConfigMap(cr)

	exporter := gitlabctl.ExporterConfigMap(cr)

	registry := gitlabctl.RegistryConfigMap(cr)

	taskRunner := gitlabctl.TaskRunnerConfigMap(cr)

	migration := gitlabctl.MigrationsConfigMap(cr)

	initdb := gitlabctl.PostgresInitDBConfigMap(cr)

	configmaps = append(configmaps,
		shell,
		gitaly,
		redis,
		redisScripts,
		webservice,
		workhorse,
		initdb,
		gitlab,
		sidekiq,
		exporter,
		registry,
		taskRunner,
		migration,
	)

	for _, cm := range configmaps {
		if err := r.createKubernetesResource(cm, cr); err != nil {
			return err
		}
	}

	return nil
}

func (r *GitLabReconciler) reconcileGitlabExporterDeployment(cr *gitlabv1beta1.GitLab) error {
	exporter := gitlabctl.ExporterDeployment(cr)

	if r.isObjectFound(exporter) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, exporter, r.Scheme); err != nil {
		return err
	}

	return r.Create(context.TODO(), exporter)
}

func (r *GitLabReconciler) reconcileJobs(cr *gitlabv1beta1.GitLab) error {

	// initialize buckets once s3 storage is up
	buckets := gitlabctl.BucketCreationJob(cr)
	if err := r.createKubernetesResource(buckets, cr); err != nil {
		return err
	}

	migration := gitlabctl.MigrationsJob(cr)
	return r.createKubernetesResource(migration, cr)
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

func (r *GitLabReconciler) reconcileDeployments(cr *gitlabv1beta1.GitLab) error {

	if err := r.reconcileWebserviceDeployment(cr); err != nil {
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

func (r *GitLabReconciler) reconcileStatefulSets(cr *gitlabv1beta1.GitLab) error {

	var statefulsets []*appsv1.StatefulSet

	postgres := gitlabctl.PostgresStatefulSet(cr)

	redis := gitlabctl.RedisStatefulSet(cr)

	gitaly := gitlabctl.GitalyStatefulSet(cr)

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

	return r.Create(context.TODO(), object.(runtime.Object))
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

func (r *GitLabReconciler) reconcileWebserviceDeployment(cr *gitlabv1beta1.GitLab) error {
	webservice := gitlabctl.WebserviceDeployment(cr)

	if r.isObjectFound(webservice) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, webservice, r.Scheme); err != nil {
		return err
	}

	return r.Create(context.TODO(), webservice)
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

func (r *GitLabReconciler) reconcileRegistryDeployment(cr *gitlabv1beta1.GitLab) error {
	registry := gitlabctl.RegistryDeployment(cr)

	if r.isObjectFound(registry) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, registry, r.Scheme); err != nil {
		return err
	}

	return r.Create(context.TODO(), registry)
}

func (r *GitLabReconciler) reconcileSecrets(cr *gitlabv1beta1.GitLab) error {
	var secrets []*corev1.Secret

	gitaly := gitlabctl.GitalySecret(cr)

	workhorse := gitlabctl.WorkhorseSecret(cr)

	registry := gitlabctl.RegistryHTTPSecret(cr)

	registryCert := gitlabctl.RegistryCertSecret(cr)

	rails := gitlabctl.RailsSecret(cr)

	postgres := gitlabctl.PostgresSecret(cr)

	redis := gitlabctl.RedisSecret(cr)

	runner := gitlabctl.RunnerRegistrationSecret(cr)

	root := gitlabctl.RootUserSecret(cr)

	smtp := gitlabctl.SMTPSettingsSecret(cr)

	shell := gitlabctl.ShellSecret(cr)

	keys := gitlabctl.ShellSSHKeysSecret(cr)

	secrets = append(secrets,
		gitaly,
		registry,
		registryCert,
		workhorse,
		rails,
		postgres,
		redis,
		root,
		runner,
		smtp,
		shell,
		keys,
	)

	for _, secret := range secrets {
		if err := r.createKubernetesResource(secret, cr); err != nil {
			return err
		}
	}

	return nil
}

func (r *GitLabReconciler) reconcileServices(cr *gitlabv1beta1.GitLab) error {
	var services []*corev1.Service

	postgres := gitlabctl.PostgresqlService(cr)

	postgresHeadless := gitlabctl.PostgresHeadlessService(cr)

	redis := gitlabctl.RedisService(cr)

	redisHeadless := gitlabctl.RedisHeadlessService(cr)

	gitaly := gitlabctl.GitalyService(cr)

	registry := gitlabctl.RegistryService(cr)

	webservice := gitlabctl.WebserviceService(cr)

	shell := gitlabctl.ShellService(cr)

	exporter := gitlabctl.ExporterService(cr)

	services = append(services,
		postgres,
		postgresHeadless,
		redis,
		redisHeadless,
		gitaly,
		registry,
		webservice,
		shell,
		exporter,
	)

	for _, svc := range services {
		if err := r.createKubernetesResource(svc, cr); err != nil {
			return err
		}
	}

	return nil
}

func (r *GitLabReconciler) reconcileShellDeployment(cr *gitlabv1beta1.GitLab) error {
	shell := gitlabctl.ShellDeployment(cr)

	if r.isObjectFound(shell) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, shell, r.Scheme); err != nil {
		return err
	}

	return r.Create(context.TODO(), shell)
}

func (r *GitLabReconciler) reconcileSidekiqDeployment(cr *gitlabv1beta1.GitLab) error {
	sidekiq := gitlabctl.SidekiqDeployment(cr)

	if r.isObjectFound(sidekiq) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, sidekiq, r.Scheme); err != nil {
		return err
	}

	return r.Create(context.TODO(), sidekiq)
}

func (r *GitLabReconciler) reconcileTaskRunnerDeployment(cr *gitlabv1beta1.GitLab) error {
	tasker := gitlabctl.TaskRunnerDeployment(cr)

	if r.isObjectFound(tasker) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, tasker, r.Scheme); err != nil {
		return err
	}

	return r.Create(context.TODO(), tasker)
}

// func (r *GitLabReconciler) reconcileRoute(cr *gitlabv1beta1.GitLab) error {
// 	workhorse := getGitlabRoute(cr)

// 	if err := r.createKubernetesResource(workhorse, cr); err != nil {
// 		return err
// 	}

// 	registry := getRegistryRoute(cr)

// 	if err := r.createKubernetesResource(registry, cr); err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (r *GitLabReconciler) reconcileIngress(cr *gitlabv1beta1.GitLab) error {
// 	controller := getIngressController(cr)
// 	if err := r.createKubernetesResource(controller, nil); err != nil {
// 		return err
// 	}

// 	var ingresses []*extensionsv1beta1.Ingress
// 	gitlab := getGitlabIngress(cr)

// 	registry := getRegistryIngress(cr)

// 	ingresses = append(ingresses,
// 		gitlab,
// 		registry,
// 	)

// 	for _, ingress := range ingresses {
// 		if err := r.createKubernetesResource(ingress, cr); err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

// func (r *GitLabReconciler) reconcileCertManagerCertificates(cr *gitlabv1beta1.GitLab) error {
// 	// certificates := RequiresCertificate(cr)

// 	issuer := CertificateIssuer(cr)

// 	return r.createKubernetesResource(issuer, cr)
// }
