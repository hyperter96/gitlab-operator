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

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
	runnerctl "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/runner"
	gitlabutils "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// RunnerReconciler reconciles a Runner object
type RunnerReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=apps.gitlab.com,resources=runners,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.gitlab.com,resources=runners/finalizers,verbs=update;patch;delete
// +kubebuilder:rbac:groups=apps.gitlab.com,resources=runners/status,verbs=get;update;patch

// Reconcile triggers when an event occurs on the watched resource
func (r *RunnerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("runner", req.NamespacedName)

	log.Info("Reconciling Runner", "name", req.NamespacedName.Name, "namespace", req.NamespacedName.Namespace)
	runner := &gitlabv1beta1.Runner{}
	if err := r.Get(ctx, req.NamespacedName, runner); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	if err := r.reconcileConfigMaps(ctx, runner); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileSecrets(ctx, runner); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileDeployments(ctx, runner); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileStatus(runner); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileMetrics(ctx, runner); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileServiceMonitor(ctx, runner); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager configures the custom resource watched resources
func (r *RunnerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gitlabv1beta1.Runner{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.Service{}).
		Owns(&appsv1.Deployment{}).
		Owns(&monitoringv1.ServiceMonitor{}).
		Complete(r)
}

func (r *RunnerReconciler) reconcileSecrets(ctx context.Context, cr *gitlabv1beta1.Runner) error {
	tokens, err := runnerctl.GetSecret(r, cr)
	if err != nil {
		return err
	}

	found := &corev1.Secret{}
	err = r.Get(ctx, types.NamespacedName{Name: tokens.Name, Namespace: cr.Namespace}, found)
	if err != nil {
		if errors.IsNotFound(err) {
			return r.Create(ctx, tokens)
		}

		return err
	}

	if reflect.DeepEqual(tokens.Data, found.Data) {
		found.Data = tokens.Data
		return r.Update(ctx, found)
	}

	return nil
}

func (r *RunnerReconciler) reconcileConfigMaps(ctx context.Context, cr *gitlabv1beta1.Runner) error {
	configs := runnerctl.GetConfigMap(cr)

	found := &corev1.ConfigMap{}
	err := r.Get(ctx, types.NamespacedName{Name: configs.Name, Namespace: cr.Namespace}, found)
	if err != nil {
		if errors.IsNotFound(err) {
			return r.Create(ctx, configs)
		}

		return err
	}

	if reflect.DeepEqual(configs.Data, found.Data) {
		found.Data = configs.Data
		return r.Update(ctx, found)
	}

	return nil
}

func (r *RunnerReconciler) reconcileDeployments(ctx context.Context, cr *gitlabv1beta1.Runner) error {
	runner := runnerctl.GetDeployment(cr)

	found := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: runner.Name, Namespace: cr.Namespace}, found)
	if err != nil {
		if errors.IsNotFound(err) {
			return r.Create(ctx, runner)
		}

		return err
	}

	if reflect.DeepEqual(runner.Spec, found.Spec) {
		found.Spec = runner.Spec
		return r.Update(ctx, found)
	}

	return nil
}

func (r *RunnerReconciler) updateRunnerStatus(cr *gitlabv1beta1.Runner, consoleLog string) error {
	runner := &gitlabv1beta1.Runner{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, runner)
	if err != nil {
		return err
	}

	if consoleLog != "" {
		runner.Status.Phase = "Running"
		runner.Status.Registration = runnerctl.RegistrationStatus(consoleLog)
	} else {
		runner.Status.Phase = "Initializing"
	}

	return r.Status().Update(context.TODO(), runner)
}

func (r *RunnerReconciler) reconcileStatus(cr *gitlabv1beta1.Runner) error {

	client, err := gitlabutils.KubernetesConfig().NewKubernetesClient()
	if err != nil {
		return err
	}

	pod, err := runnerctl.WorkerPod(cr, client)
	if err != nil {
		return err
	}

	var log string
	if gitlabutils.IsPodRunning(pod) {
		log, err = runnerctl.LogStream(pod, client)
		if err != nil {
			return err
		}
	}

	if err := r.updateRunnerStatus(cr, log); err != nil {
		return err
	}

	return nil
}

func (r *RunnerReconciler) reconcileMetrics(ctx context.Context, cr *gitlabv1beta1.Runner) error {
	svc := runnerctl.MetricsService(cr)

	found := &corev1.Service{}
	err := r.Get(ctx, types.NamespacedName{Name: svc.Name, Namespace: cr.Namespace}, found)
	if err != nil {
		if errors.IsNotFound(err) {
			return r.Create(ctx, svc)
		}

		return err
	}

	if reflect.DeepEqual(svc.Spec, found.Spec) {
		found.Spec = svc.Spec
		return r.Update(ctx, found)
	}

	return nil
}

func (r *RunnerReconciler) reconcileServiceMonitor(ctx context.Context, cr *gitlabv1beta1.Runner) error {

	if gitlabutils.IsPrometheusSupported() {
		sm := runnerctl.ServiceMonitorService(cr)

		found := &monitoringv1.ServiceMonitor{}
		err := r.Get(ctx, types.NamespacedName{Name: sm.Name, Namespace: cr.Namespace}, found)
		if err != nil {
			if errors.IsNotFound(err) {
				return r.Create(ctx, sm)
			}

			return err
		}

		if reflect.DeepEqual(sm.Spec, found.Spec) {
			found.Spec = sm.Spec
			return r.Update(ctx, found)
		}
	}

	return nil
}
