package gitlab

import (
	"context"
	"sync"
	"time"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	certmanagerv1alpha2 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	nginxv1alpha1 "github.com/nginxinc/nginx-ingress-operator/pkg/apis/k8s/v1alpha1"
	routev1 "github.com/openshift/api/route/v1"
	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/pkg/apis/gitlab/v1beta1"
	gitlabutils "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/pkg/controller/utils"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
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

	err = c.Watch(&source.Kind{Type: &batchv1.Job{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &gitlabv1beta1.Gitlab{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &certmanagerv1alpha2.Issuer{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &gitlabv1beta1.Gitlab{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &certmanagerv1alpha2.Certificate{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &gitlabv1beta1.Gitlab{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &nginxv1alpha1.NginxIngressController{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &gitlabv1beta1.Gitlab{},
	})
	if err != nil {
		return err
	}

	if gitlabutils.IsOpenshift() {
		// Watch Openshifts route if running on Openshift
		err = c.Watch(&source.Kind{Type: &routev1.Route{}}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &gitlabv1beta1.Gitlab{},
		})
		if err != nil {
			return err
		}
	} else {
		// Watch Ingress resources in a Kubernetes environment
		err = c.Watch(&source.Kind{Type: &extensionsv1beta1.Ingress{}}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &gitlabv1beta1.Gitlab{},
		})
		if err != nil {
			return err
		}
	}

	if gitlabutils.IsPrometheusSupported() {
		// Watch Openshifts route if running on Openshift
		err = c.Watch(&source.Kind{Type: &monitoringv1.ServiceMonitor{}}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &gitlabv1beta1.Gitlab{},
		})
		if err != nil {
			return err
		}
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

// Reconcile child resources used by the operator
func (r *ReconcileGitlab) reconcileChildResources(cr *gitlabv1beta1.Gitlab) error {

	if err := r.reconcileConfigMaps(cr); err != nil {
		return err
	}

	if err := r.reconcileSecrets(cr); err != nil {
		return err
	}

	if err := r.reconcileServices(cr); err != nil {
		return err
	}

	if err := r.maskEmailPasword(cr); err != nil {
		return err
	}

	if err := r.reconcileStatefulSets(cr); err != nil {
		return err
	}

	if err := r.reconcileMinioInstance(cr); err != nil {
		return err
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		for !isEndpointReady(cr.Name+"-postgresql", cr) {
			time.Sleep(time.Second * 1)
		}
		wg.Done()
	}()

	if RequiresCertManagerCertificate(cr).All() {
		if err := r.reconcileCertManagerCertificates(cr); err != nil {
			return err
		}
	}

	wg.Wait()

	if err := r.reconcileJobs(cr); err != nil {
		return err
	}

	if err := r.reconcileDeployments(cr); err != nil {
		return err
	}

	// Deploy ingress to expose GitLab
	if err := r.reconcileIngress(cr); err != nil {
		return err
	}

	if gitlabutils.IsPrometheusSupported() {
		// Deploy a prometheus service monitor
		if err := r.reconcileServiceMonitor(cr); err != nil {
			return err
		}
	}

	if err := r.reconcileGitlabStatus(cr); err != nil {
		return err
	}

	return nil
}
