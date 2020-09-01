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
	"reflect"
	"strings"

	"github.com/go-logr/logr"
	"github.com/prometheus/common/log"
	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/backup"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
	gitlabutils "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/utils"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GLBackupReconciler reconciles a GLBackup object
type GLBackupReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=apps.gitlab.com,namespace="placeholder",resources=glbackups,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.gitlab.com,namespace="placeholder",resources=glbackups/status,verbs=get;update;patch

// Reconcile triggers when an event occurs on the watched resource
func (r *GLBackupReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("glbackup", req.NamespacedName)

	backup := &gitlabv1beta1.GLBackup{}
	if err := r.Get(ctx, req.NamespacedName, backup); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	if err := r.reconcileBackup(ctx, backup); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileBackupConfigMap(ctx, backup); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager configures the custom resource watched resources
func (r *GLBackupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gitlabv1beta1.GLBackup{}).
		Complete(r)
}

func (r *GLBackupReconciler) reconcileBackup(ctx context.Context, cr *gitlabv1beta1.GLBackup) error {
	if cr.Spec.Schedule != "" &&
		len(strings.Split(cr.Spec.Schedule, " ")) == 5 {
		return r.setupBackupSchedule(ctx, cr)
	}

	return r.reconcileBackupJob(ctx, cr)
}

func (r *GLBackupReconciler) setupBackupSchedule(ctx context.Context, cr *gitlabv1beta1.GLBackup) error {
	backup := backup.NewSchedule(cr)

	found := &batchv1beta1.CronJob{}
	err := r.Get(ctx, types.NamespacedName{Name: backup.Name, Namespace: cr.Namespace}, found)
	if err != nil {
		if errors.IsNotFound(err) {
			return r.Create(ctx, backup)
		}

		return err
	}

	// if stored and generated schedule do not match
	if !reflect.DeepEqual(backup.Spec, found.Spec) {
		return r.Update(ctx, found)
	}

	return nil
}

func (r *GLBackupReconciler) reconcileBackupJob(ctx context.Context, cr *gitlabv1beta1.GLBackup) error {
	backup := backup.NewBackup(cr)

	if r.IfObjectExists(types.NamespacedName{Name: backup.Name, Namespace: cr.Namespace}, backup) {
		return r.updateBackupJob(backup)
	}

	return r.createKubernetesResource(backup, cr)
}

func (r *GLBackupReconciler) updateBackupJob(backup *batchv1.Job) error {
	found := &batchv1.Job{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: backup.Name, Namespace: backup.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		return err
	}

	if !reflect.DeepEqual(backup.Spec, found.Spec) {
		log.Info("The specs do not match")
	}

	return nil
}

func (r *GLBackupReconciler) reconcileBackupSchedule(cr *gitlabv1beta1.GLBackup) error {
	backup := backup.NewSchedule(cr)

	if r.IfObjectExists(types.NamespacedName{Name: backup.Name, Namespace: cr.Namespace}, backup) {
		return r.updateBackupSchedule(backup)
	}

	return r.createKubernetesResource(backup, cr)
}

func (r *GLBackupReconciler) updateBackupSchedule(backup *batchv1beta1.CronJob) error {
	found := &batchv1beta1.CronJob{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: backup.Name, Namespace: backup.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		return err
	}

	if found.Spec.Schedule != backup.Spec.Schedule {
		found.Spec.Schedule = backup.Spec.Schedule
	}

	return r.Update(context.TODO(), backup)
}

func (r *GLBackupReconciler) reconcileBackupConfigMap(ctx context.Context, cr *gitlabv1beta1.GLBackup) error {
	backupLock := backup.LockConfigMap(cr)

	if r.IfObjectExists(types.NamespacedName{Name: backupLock.Name, Namespace: cr.Namespace}, backupLock) {
		if err := r.reconcileBackupStatus(cr); err != nil {
			return err
		}

		lock := &corev1.ConfigMap{}
		if err := r.Get(context.TODO(), types.NamespacedName{Name: backupLock.Name, Namespace: cr.Namespace}, lock); err != nil {
			return err
		}
		lock.Data = map[string]string{}

		return r.Patch(ctx, lock, client.MergeFrom(backupLock))
	}

	return r.createKubernetesResource(backupLock, cr)
}

func (r *GLBackupReconciler) createKubernetesResource(object interface{}, parent *gitlabv1beta1.GLBackup) error {

	// If parent resource is not nil, owner reference will be set
	if parent != nil {
		if err := controllerutil.SetControllerReference(parent, object.(metav1.Object), r.Scheme); err != nil {
			return err
		}
	}

	return r.Create(context.TODO(), object.(runtime.Object))
}

// IfObjectExists returns true if a given kubernetes object exists
func (r *GLBackupReconciler) IfObjectExists(key types.NamespacedName, result runtime.Object) bool {
	err := r.Get(context.TODO(), key, result)
	if err != nil && errors.IsNotFound(err) {
		return false
	}

	return true
}

// TODO: update backup status implementation
func (r *GLBackupReconciler) reconcileBackupStatus(cr *gitlabv1beta1.GLBackup) error {
	lockName := strings.Join([]string{cr.Name, "backup", "lock"}, "-")
	backupData, err := gitlabutils.ConfigMapData(lockName, cr.Namespace)
	if err != nil {
		log.Error(err, "Error getting configmap data")
	}

	backup := &gitlabv1beta1.GLBackup{}
	if err := r.Get(context.TODO(), types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, backup); err != nil {
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
		return r.Status().Update(context.TODO(), backup)
	}

	return nil
}

func (r *GLBackupReconciler) getBackupJobResource(cr *gitlabv1beta1.GLBackup) *batchv1.Job {
	job := &batchv1.Job{}
	jobName := strings.Join([]string{cr.Name, "backup"}, "-")
	err := r.Get(context.TODO(), types.NamespacedName{Name: jobName, Namespace: cr.Namespace}, job)
	if err != nil && errors.IsNotFound(err) {
		return nil
	}

	return job
}

func (r *GLBackupReconciler) isBackupFailed(cr *gitlabv1beta1.GLBackup, backupError string) bool {
	job := r.getBackupJobResource(cr)
	if job == nil {
		return false
	}

	return job.Status.Succeeded < 1 && !strings.Contains(backupError, "Module python-magic is not available")
}

func (r *GLBackupReconciler) isBackupRunning(cr *gitlabv1beta1.GLBackup, data map[string]string) bool {

	job := r.getBackupJobResource(cr)
	if job == nil {
		return false
	}

	return job.Status.CompletionTime == nil && cr.Spec.Schedule == "" && len(data) == 0
}

func (r *GLBackupReconciler) isBackupComplete(cr *gitlabv1beta1.GLBackup, backupOutput string) bool {
	return cr.Status.CompletedAt != "" // && backupOutput != ""
}
