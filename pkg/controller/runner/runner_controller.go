package runner

import (
	"context"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	gitlabv1beta1 "gitlab.com/ochienged/gitlab-operator/pkg/apis/gitlab/v1beta1"
	gitlabutils "gitlab.com/ochienged/gitlab-operator/pkg/controller/utils"
	appsv1 "k8s.io/api/apps/v1"
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

var log = logf.Log.WithName("controller_runner")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Runner Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileRunner{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("runner-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Runner
	err = c.Watch(&source.Kind{Type: &gitlabv1beta1.Runner{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Runner
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &gitlabv1beta1.Runner{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &gitlabv1beta1.Runner{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &gitlabv1beta1.Runner{},
	})
	if err != nil {
		return err
	}

	if gitlabutils.IsPrometheusSupported() {
		err = c.Watch(&source.Kind{Type: &monitoringv1.ServiceMonitor{}}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &gitlabv1beta1.Runner{},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// blank assignment to verify that ReconcileRunner implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileRunner{}

// ReconcileRunner reconciles a Runner object
type ReconcileRunner struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Runner object and makes changes based on the state read
// and what is in the Runner.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileRunner) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Runner")

	// Fetch the Runner instance
	instance := &gitlabv1beta1.Runner{}
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

	if err := r.reconcileResources(instance); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileRunner) reconcileResources(cr *gitlabv1beta1.Runner) (err error) {
	if err = r.reconcileConfigMap(cr); err != nil {
		return
	}

	if err = r.reconcileSecrets(cr); err != nil {
		return
	}

	if err = r.reconcileDeployments(cr); err != nil {
		return
	}

	if err = r.reconcileRunnerStatus(cr); err != nil {
		return
	}

	if err = r.reconcileRunnerMetrics(cr); err != nil {
		return
	}

	return
}

func (r *ReconcileRunner) reconcileSecrets(cr *gitlabv1beta1.Runner) error {
	tokens := getRunnerSecret(r.client, cr)

	if err := r.createKubernetesResource(cr, tokens); err != nil {
		return err
	}

	return nil
}

func (r *ReconcileRunner) reconcileConfigMap(cr *gitlabv1beta1.Runner) error {
	configs := getRunnerConfigMap(cr)

	if err := r.createKubernetesResource(cr, configs); err != nil {
		return err
	}

	return nil
}

func (r *ReconcileRunner) reconcileDeployments(cr *gitlabv1beta1.Runner) error {
	runner := getRunnerDeployment(cr)

	if err := r.createKubernetesResource(cr, runner); err != nil {
		return err
	}

	return nil
}

func objectNamespacedName(obj interface{}) types.NamespacedName {
	object := obj.(metav1.Object)
	return types.NamespacedName{Name: object.GetName(), Namespace: object.GetNamespace()}
}

func (r *ReconcileRunner) createKubernetesResource(cr *gitlabv1beta1.Runner, object interface{}) error {

	if gitlabutils.IsObjectFound(r.client, objectNamespacedName(object), object.(runtime.Object)) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, object.(metav1.Object), r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), object.(runtime.Object))
}
