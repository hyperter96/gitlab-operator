package backup

import (
	"context"
	"reflect"
	"strings"

	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/pkg/apis/gitlab/v1beta1"
	gitlabutils "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/pkg/controller/utils"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

var log = logf.Log.WithName("controller_backup")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Backup Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileBackup{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("backup-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Backup
	err = c.Watch(&source.Kind{Type: &gitlabv1beta1.Backup{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Backup
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &gitlabv1beta1.Backup{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &batchv1.Job{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &gitlabv1beta1.Backup{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &batchv1beta1.CronJob{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &gitlabv1beta1.Backup{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.ConfigMap{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &gitlabv1beta1.Backup{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileBackup implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileBackup{}

// ReconcileBackup reconciles a Backup object
type ReconcileBackup struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Backup object and makes changes based on the state read
// and what is in the Backup.Spec
func (r *ReconcileBackup) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Backup")

	// Fetch the Backup instance
	instance := &gitlabv1beta1.Backup{}
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

	if err := r.reconcileBackupResources(instance); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileBackup) createKubernetesResource(object interface{}, parent *gitlabv1beta1.Backup) error {

	// If parent resource is nil, not owner reference will be set
	if parent != nil {
		if err := controllerutil.SetControllerReference(parent, object.(metav1.Object), r.scheme); err != nil {
			return err
		}
	}

	return r.client.Create(context.TODO(), object.(runtime.Object))
}

func (r *ReconcileBackup) reconcileBackupResources(cr *gitlabv1beta1.Backup) error {

	if err := r.reconcileBackup(cr); err != nil {
		return err
	}

	if err := r.reconcileBackupConfigMap(cr); err != nil {
		return err
	}

	return nil
}

func (r *ReconcileBackup) reconcileBackup(cr *gitlabv1beta1.Backup) error {

	if cr.Spec.Schedule != "" &&
		len(strings.Split(cr.Spec.Schedule, " ")) == 5 {
		return r.reconcileBackupSchedule(cr)
	}

	return r.reconcileBackupJob(cr)
}

func (r *ReconcileBackup) reconcileBackupJob(cr *gitlabv1beta1.Backup) error {
	backup := NewBackup(cr)

	if r.IfObjectExists(types.NamespacedName{Name: backup.Name, Namespace: cr.Namespace}, backup) {
		return r.updateBackupJob(backup)
	}

	return r.createKubernetesResource(backup, cr)
}

func (r *ReconcileBackup) reconcileBackupSchedule(cr *gitlabv1beta1.Backup) error {
	backup := NewBackupSchedule(cr)

	if r.IfObjectExists(types.NamespacedName{Name: backup.Name, Namespace: cr.Namespace}, backup) {
		return r.updateBackupSchedule(backup)
	}

	return r.createKubernetesResource(backup, cr)
}

// UpdateBackupJob updates existing the kubernetes job
func (r *ReconcileBackup) updateBackupJob(backup *batchv1.Job) error {
	found := &batchv1.Job{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: backup.Name, Namespace: backup.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		return err
	}

	if !reflect.DeepEqual(backup.Spec, found.Spec) {
		log.Info("The specs do not match")
	}

	return nil
}

func (r *ReconcileBackup) updateBackupSchedule(backup *batchv1beta1.CronJob) error {
	found := &batchv1beta1.CronJob{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: backup.Name, Namespace: backup.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		return err
	}

	if found.Spec.Schedule != backup.Spec.Schedule {
		found.Spec.Schedule = backup.Spec.Schedule
	}

	return r.client.Update(context.TODO(), backup)
}

// IfObjectExists returns true if a given kubernetes object exists
func (r *ReconcileBackup) IfObjectExists(key types.NamespacedName, result runtime.Object) bool {
	err := r.client.Get(context.TODO(), key, result)
	if err != nil && errors.IsNotFound(err) {
		return false
	}

	return true
}

func (r *ReconcileBackup) reconcileBackupConfigMap(cr *gitlabv1beta1.Backup) error {
	backupLock := backupConfigMap(cr)

	if r.IfObjectExists(types.NamespacedName{Name: backupLock.Name, Namespace: cr.Namespace}, backupLock) {
		if err := r.reconcileBackupStatus(cr); err != nil {
			return err
		}

		lock := &corev1.ConfigMap{}
		if err := r.client.Get(context.TODO(), types.NamespacedName{Name: backupLock.Name, Namespace: cr.Namespace}, lock); err != nil {
			return err
		}
		lock.Data = map[string]string{}

		return r.client.Patch(context.TODO(), lock, client.MergeFrom(backupLock))
	}

	return r.createKubernetesResource(backupLock, cr)
}

func (r *ReconcileBackup) reconcileBackupStatus(cr *gitlabv1beta1.Backup) error {
	lockName := strings.Join([]string{cr.Name, "backup", "lock"}, "-")
	backupData, err := gitlabutils.ConfigMapData(lockName, cr.Namespace)
	if err != nil {
		log.Error(err, "Error getting configmap data")
	}

	backup := &gitlabv1beta1.Backup{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, backup); err != nil {
	}

	if start, ok := backupData["startTime"]; ok {
		backup.Status.StartedAt = start
	}

	if completed, ok := backupData["stopTime"]; ok {
		backup.Status.CompletedAt = completed
	}

	if r.isBackupRunning(cr, backupData) && backup.Status.CompletedAt == "" {
		backup.Status.Phase = gitlabv1beta1.BackupRunning
	}

	if bkOut, ok := backupData["output"]; ok {
		if r.isBackupComplete(backup, bkOut) {
			backup.Status.Phase = gitlabv1beta1.BackupCompleted
		}
	}

	if bkErr, ok := backupData["error"]; ok {
		if r.isBackupFailed(cr, bkErr) {
			backup.Status.Phase = gitlabv1beta1.BackupFailed
		}
	}

	if backup.Spec.Schedule != "" {
		backup.Status.Phase = gitlabv1beta1.BackupScheduled
	}

	if !reflect.DeepEqual(cr.Status, backup.Status) {
		return r.client.Status().Update(context.TODO(), backup)
	}

	return nil
}

func (r *ReconcileBackup) getBackupJobResource(cr *gitlabv1beta1.Backup) *batchv1.Job {
	job := &batchv1.Job{}
	jobName := strings.Join([]string{cr.Name, "backup"}, "-")
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: jobName, Namespace: cr.Namespace}, job)
	if err != nil && errors.IsNotFound(err) {
		return nil
	}

	return job
}

func (r *ReconcileBackup) isBackupFailed(cr *gitlabv1beta1.Backup, backupError string) bool {
	job := r.getBackupJobResource(cr)
	if job == nil {
		return false
	}

	return job.Status.Succeeded < 1 && !strings.Contains(backupError, "Module python-magic is not available")
}

func (r *ReconcileBackup) isBackupRunning(cr *gitlabv1beta1.Backup, data map[string]string) bool {

	job := r.getBackupJobResource(cr)
	if job == nil {
		return false
	}

	return job.Status.CompletionTime == nil && cr.Spec.Schedule == "" && len(data) == 0
}

func (r *ReconcileBackup) isBackupComplete(cr *gitlabv1beta1.Backup, backupOutput string) bool {
	return cr.Status.CompletedAt != "" // && backupOutput != ""
}
