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

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
	runnercontroller "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/runner"
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

// +kubebuilder:rbac:groups=apps.gitlab.com,namespace="placeholder",resources=runners,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.gitlab.com,namespace="placeholder",resources=runners/status,verbs=get;update;patch

// Reconcile triggers when an event occurs on the watched resource
func (r *RunnerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("runner", req.NamespacedName)

	runner := &gitlabv1beta1.Runner{}
	if err := r.Get(ctx, req.NamespacedName, runner); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	if err := r.reconcileConfigMaps(runner); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileSecrets(runner); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileDeployments(runner); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileStatus(runner); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileMetrics(runner); err != nil {
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

func (r *RunnerReconciler) reconcileSecrets(cr *gitlabv1beta1.Runner) error {
	tokens := runnercontroller.GetSecret(r, cr)

	if err := r.createKubernetesResource(cr, tokens); err != nil {
		return err
	}

	return nil
}

func (r *RunnerReconciler) reconcileConfigMaps(cr *gitlabv1beta1.Runner) error {
	configs := runnercontroller.GetConfigMap(cr)

	if err := r.createKubernetesResource(cr, configs); err != nil {
		return err
	}

	return nil
}

func (r *RunnerReconciler) reconcileDeployments(cr *gitlabv1beta1.Runner) error {
	runner := runnercontroller.GetDeployment(cr)

	if err := r.createKubernetesResource(cr, runner); err != nil {
		return err
	}

	return nil
}

func (r *RunnerReconciler) createKubernetesResource(cr *gitlabv1beta1.Runner, object interface{}) error {

	obj := object.(metav1.Object)
	nsName := types.NamespacedName{Name: obj.GetNamespace(), Namespace: obj.GetNamespace()}

	if gitlabutils.IsObjectFound(r, nsName, object.(runtime.Object)) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, object.(metav1.Object), r.Scheme); err != nil {
		return err
	}

	return r.Create(context.TODO(), object.(runtime.Object))
}

func (r *RunnerReconciler) updateRunnerStatus(cr *gitlabv1beta1.Runner, consoleLog string) error {
	runner := &gitlabv1beta1.Runner{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, runner)
	if err != nil {
		return err
	}

	if consoleLog != "" {
		runner.Status.Phase = "Running"
		runner.Status.Registration = runnercontroller.RegistrationStatus(consoleLog)
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

	pod, err := runnercontroller.WorkerPod(cr, client)
	if err != nil {
		return err
	}

	var log string
	if gitlabutils.IsPodRunning(pod) {
		log, err = runnercontroller.LogStream(pod, client)
		if err != nil {
			return err
		}
	}

	if err := r.updateRunnerStatus(cr, log); err != nil {
		return err
	}

	return nil
}

func (r *RunnerReconciler) reconcileMetrics(cr *gitlabv1beta1.Runner) error {
	svc := runnercontroller.MetricsService(cr)

	if err := r.createKubernetesResource(cr, svc); err != nil {
		return err
	}

	if gitlabutils.IsPrometheusSupported() {
		sm := runnercontroller.ServiceMonitorService(cr)

		if err := r.createKubernetesResource(cr, sm); err != nil {
			return err
		}
	}

	return nil
}
