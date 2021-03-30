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

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/internal"
	runnerctl "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/runner"
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
// +kubebuilder:rbac:groups=core,resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods/log,verbs=get

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

	if err := r.reconcileServiceAccount(ctx, runner); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileConfigMaps(ctx, runner); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.validateRegistrationTokenSecret(ctx, runner); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileDeployments(ctx, runner, log); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileStatus(ctx, runner); err != nil {
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

func (r *RunnerReconciler) reconcileConfigMaps(ctx context.Context, cr *gitlabv1beta1.Runner) error {
	configs := runnerctl.ConfigMap(cr)

	// load user provided config.toml
	if cr.Spec.Configuration != "" {
		userToml, err := r.getUserConfigToml(ctx, cr)
		if err != nil {
			return err
		}

		configs.Data["config.toml"] = userToml
	}

	found := &corev1.ConfigMap{}
	err := r.Get(ctx, types.NamespacedName{Name: configs.Name, Namespace: cr.Namespace}, found)
	if err != nil {
		if errors.IsNotFound(err) {
			if err := controllerutil.SetControllerReference(cr, configs, r.Scheme); err != nil {
				return err
			}

			return r.Create(ctx, configs)
		}

		return err
	}

	if !reflect.DeepEqual(configs.Data, found.Data) {
		found.Data = configs.Data
		return r.Update(ctx, found)
	}

	return nil
}

func (r *RunnerReconciler) getUserConfigToml(ctx context.Context, cr *gitlabv1beta1.Runner) (string, error) {
	userCM := &corev1.ConfigMap{}
	userCMKey := types.NamespacedName{
		Name:      cr.Spec.Configuration,
		Namespace: cr.Namespace,
	}

	if err := r.Get(ctx, userCMKey, userCM); err != nil {
		return "", err
	}

	userToml, ok := userCM.Data["config.toml"]
	if ok {
		return userToml, nil
	}

	return "", fmt.Errorf("config.toml key not found")
}

func (r *RunnerReconciler) reconcileDeployments(ctx context.Context, cr *gitlabv1beta1.Runner, log logr.Logger) error {
	runner := runnerctl.Deployment(cr)

	if err := r.appendConfigMapChecksum(ctx, runner); err != nil {
		log.Error(err, "Error appending configmap checksums")
	}

	found := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: runner.Name, Namespace: cr.Namespace}, found)
	if err != nil {
		if errors.IsNotFound(err) {
			if err := controllerutil.SetControllerReference(cr, runner, r.Scheme); err != nil {
				return err
			}

			return r.Create(ctx, runner)
		}

		return err
	}

	deployment, changed := internal.IsDeploymentChanged(found, runner)
	if changed {
		return r.Update(ctx, deployment)
	}

	return nil
}

func (r *RunnerReconciler) updateRunnerStatus(ctx context.Context, cr *gitlabv1beta1.Runner, consoleLog string) error {
	runner := &gitlabv1beta1.Runner{}
	err := r.Get(ctx, types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, runner)
	if err != nil {
		return err
	}

	if consoleLog != "" {
		runner.Status.Phase = "Running"
		runner.Status.Registration = runnerctl.RegistrationStatus(consoleLog)
	} else {
		runner.Status.Phase = "Initializing"
	}

	return r.Status().Update(ctx, runner)
}

func (r *RunnerReconciler) reconcileStatus(ctx context.Context, cr *gitlabv1beta1.Runner) error {
	// set := labels.Set(map[string]string{})
	// selector := client.MatchingLabelsSelector{
	// 	Selector: labels.SelectorFromSet(set),
	// }

	// podlist := &corev1.PodList{}
	// err := r.List(ctx, podlist, client.InNamespace(cr.Namespace), selector)
	// if err != nil {
	// }

	client, err := internal.KubernetesConfig().NewKubernetesClient()
	if err != nil {
		return err
	}

	pod, err := runnerctl.WorkerPod(cr, client)
	if err != nil {
		return err
	}

	var log string
	if internal.IsPodRunning(pod) {
		log, err = runnerctl.LogStream(pod, client)
		if err != nil {
			return err
		}
	}

	if err := r.updateRunnerStatus(ctx, cr, log); err != nil {
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
			if err := controllerutil.SetControllerReference(cr, svc, r.Scheme); err != nil {
				return err
			}

			return r.Create(ctx, svc)
		}

		return err
	}

	if !reflect.DeepEqual(svc.Spec, found.Spec) {
		// besides ClusterIP, not much is expected to change
		// return r.Update(ctx, found)
		return nil
	}

	return nil
}

func (r *RunnerReconciler) reconcileServiceMonitor(ctx context.Context, cr *gitlabv1beta1.Runner) error {

	if internal.IsPrometheusSupported() {
		sm := runnerctl.ServiceMonitorService(cr)

		found := &monitoringv1.ServiceMonitor{}
		err := r.Get(ctx, types.NamespacedName{Name: sm.Name, Namespace: cr.Namespace}, found)
		if err != nil {
			if errors.IsNotFound(err) {
				if err := controllerutil.SetControllerReference(cr, sm, r.Scheme); err != nil {
					return err
				}

				return r.Create(ctx, sm)
			}

			return err
		}

		if !reflect.DeepEqual(sm.Spec, found.Spec) {
			found.Spec = sm.Spec
			return r.Update(ctx, found)
		}
	}

	return nil
}

func (r *RunnerReconciler) validateRegistrationTokenSecret(ctx context.Context, cr *gitlabv1beta1.Runner) error {
	tokenSecretName := runnerctl.RegistrationTokenSecretName(cr)

	found := &corev1.Secret{}
	err := r.Get(ctx, types.NamespacedName{Name: tokenSecretName, Namespace: cr.Namespace}, found)
	if err != nil {
		return err
	}

	registrationToken, ok := found.Data["runner-registration-token"]
	if !ok {
		return fmt.Errorf("runner-registration-token key not found in %s secret", tokenSecretName)
	}

	tokenStr := string(registrationToken)
	if tokenStr == "" {
		return fmt.Errorf("runner-registration-token can not be empty")
	}

	if _, ok := found.StringData["runner-token"]; !ok {
		found.Data["runner-token"] = []byte("")
		return r.Update(ctx, found)
	}

	return nil
}

func (r *RunnerReconciler) appendConfigMapChecksum(ctx context.Context, deployment *appsv1.Deployment) error {
	configmaps := internal.DeploymentConfigMaps(deployment)

	for _, cmName := range configmaps {
		found := &corev1.ConfigMap{}
		err := r.Get(ctx, types.NamespacedName{Name: cmName, Namespace: deployment.Namespace}, found)
		if err != nil {
			return err
		}

		// get checksum from the configmap annotation
		if checksum, ok := found.Annotations["checksum"]; ok {
			// compare the checksum with cm checksum in deployment template annotation
			if val, ok := deployment.Spec.Template.Annotations[cmName]; ok {
				if val != checksum {
					deployment.Spec.Template.Annotations[cmName] = checksum
				}
			} else {
				// account for nil map exception
				if deployment.Spec.Template.Annotations != nil {
					deployment.Spec.Template.Annotations[cmName] = checksum
				} else {
					deployment.Spec.Template.Annotations = map[string]string{
						cmName: checksum,
					}
				}
			}
		}
	}

	return nil
}

func (r *RunnerReconciler) reconcileServiceAccount(ctx context.Context, cr *gitlabv1beta1.Runner) error {
	sa := internal.ServiceAccount("gitlab-runner", cr.Namespace)
	lookupKey := types.NamespacedName{
		Name:      sa.Name,
		Namespace: cr.Namespace,
	}

	found := &corev1.ServiceAccount{}
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
