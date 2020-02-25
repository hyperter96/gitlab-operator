package gitlab

import (
	"context"

	gitlabv1beta1 "github.com/OchiengEd/gitlab-operator/pkg/apis/gitlab/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_gitlab")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Gitlab Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileGitlab{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("gitlab-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Gitlab
	err = c.Watch(&source.Kind{Type: &gitlabv1beta1.Gitlab{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Gitlab
	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &gitlabv1beta1.Gitlab{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.ConfigMap{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &gitlabv1beta1.Gitlab{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &gitlabv1beta1.Gitlab{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.PersistentVolumeClaim{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &gitlabv1beta1.Gitlab{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &gitlabv1beta1.Gitlab{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &appsv1.StatefulSet{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &gitlabv1beta1.Gitlab{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileGitlab implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileGitlab{}

// ReconcileGitlab reconciles a Gitlab object
type ReconcileGitlab struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Gitlab object and makes changes based on the state read
// and what is in the Gitlab.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileGitlab) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Gitlab")

	// Fetch the Gitlab instance
	instance := &gitlabv1beta1.Gitlab{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if err := r.reconcileChildResources(instance); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileGitlab) isObjectFound(key types.NamespacedName, object runtime.Object) bool {
	if err := r.client.Get(context.TODO(), key, object); err != nil {
		return false
	}

	return true
}

// Reconcile child resources used by the operator
func (r *ReconcileGitlab) reconcileChildResources(cr *gitlabv1beta1.Gitlab) error {
	if err := r.reconcileSecrets(cr); err != nil {
		return err
	}

	if err := r.reconcileConfigMaps(cr); err != nil {
		return err
	}

	if err := r.reconcileServices(cr); err != nil {
		return err
	}

	if err := r.reconcilePersistentVolumeClaims(cr); err != nil {
		return err
	}

	if err := r.reconcileStatefulSets(cr); err != nil {
		return err
	}

	if err := r.reconcileDeployments(cr); err != nil {
		return err
	}

	if err := r.reconcileIngress(cr); err != nil {
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

func (r *ReconcileGitlab) reconcileSecrets(cr *gitlabv1beta1.Gitlab) error {
	core := getGilabSecret(cr)

	if r.isObjectFound(types.NamespacedName{Name: core.Name, Namespace: core.Namespace}, core) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, core, r.scheme); err != nil {
		return err
	}

	if err := r.client.Create(context.TODO(), core); err != nil {
		return err
	}

	runner := getGilabRunnerSecret(cr)

	if r.isObjectFound(types.NamespacedName{Name: runner.Name, Namespace: runner.Namespace}, runner) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, runner, r.scheme); err != nil {
		return err
	}

	if err := r.client.Create(context.TODO(), runner); err != nil {
		return err
	}

	return nil
}

func (r *ReconcileGitlab) reconcileConfigMaps(cr *gitlabv1beta1.Gitlab) error {
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

	redis := getRedisConfig(cr)

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

	gitlabRunner := getGitlabRunnerConfig(cr)

	if r.isObjectFound(types.NamespacedName{Name: gitlabRunner.Name, Namespace: gitlabRunner.Namespace}, gitlabRunner) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, gitlabRunner, r.scheme); err != nil {
		return err
	}

	if err := r.client.Create(context.TODO(), gitlabRunner); err != nil {
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

func (r *ReconcileGitlab) reconcileIngress(cr *gitlabv1beta1.Gitlab) error {
	ingress := getGitlabIngress(cr)

	if r.isObjectFound(types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, ingress) {
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
