package helm

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2beta1 "k8s.io/api/autoscaling/v2beta1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	networkingv1 "k8s.io/api/networking/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// TrueSelector is an ObjectSelector that selects all objects.
var TrueSelector = func(_ runtime.Object) bool {
	return true
}

// FalseSelector is an ObjectSelector that selects no object.
var FalseSelector = func(_ runtime.Object) bool {
	return false
}

type typeMistmatchError struct {
	expected interface{}
	observed interface{}
}

func (e *typeMistmatchError) Error() string {
	return fmt.Sprintf("expected %T, got %T", e.expected, e.observed)
}

func newTypeMistmatchError(expected, observed interface{}) error {
	return &typeMistmatchError{
		expected: expected,
		observed: observed,
	}
}

// IsTypeMistmatchError returns true if the error is raised because of type mistmatch.
func IsTypeMistmatchError(err error) bool {
	_, ok := err.(*typeMistmatchError)
	return ok
}

// ConfigMapSelector is an ObjectSelector that selects ConfigMap objects.
var ConfigMapSelector = func(o runtime.Object) bool {
	_, ok := o.(*corev1.ConfigMap)
	return ok
}

// ClusterRoleSelector is an ObjectSelector that selects ClusterRole objects.
var ClusterRoleSelector = func(o runtime.Object) bool {
	_, ok := o.(*rbacv1.ClusterRole)
	return ok
}

// ClusterRoleBindingSelector is an ObjectSelector that selects ClusterRoleBinding objects.
var ClusterRoleBindingSelector = func(o runtime.Object) bool {
	_, ok := o.(*rbacv1.ClusterRoleBinding)
	return ok
}

// CronJobSelector is an ObjectSelector that selects CronJob objects.
var CronJobSelector = func(o runtime.Object) bool {
	_, ok := o.(*batchv1beta1.CronJob)
	return ok
}

// DaemonSetSelector is an ObjectSelector that selects DaemonSet objects.
var DaemonSetSelector = func(o runtime.Object) bool {
	_, ok := o.(*appsv1.DaemonSet)
	return ok
}

// DeploymentSelector is an ObjectSelector that selects Deployment objects.
var DeploymentSelector = func(o runtime.Object) bool {
	_, ok := o.(*appsv1.Deployment)
	return ok
}

// HorizontalPodAutoscalerSelector is an ObjectSelector that selects HorizontalPodAutoscaler objects.
var HorizontalPodAutoscalerSelector = func(o runtime.Object) bool {
	_, ok := o.(*autoscalingv2beta1.HorizontalPodAutoscaler)
	return ok
}

// IngressSelector is an ObjectSelector that selects Ingress objects.
var IngressSelector = func(o runtime.Object) bool {
	_, ok := o.(*extensionsv1beta1.Ingress)
	if !ok {
		_, ok = o.(*networkingv1beta1.Ingress)
	}
	return ok
}

// JobSelector is an ObjectSelector that selects Job objects.
var JobSelector = func(o runtime.Object) bool {
	_, ok := o.(*batchv1.Job)
	return ok
}

// NetworkPolicySelector is an ObjectSelector that selects NetworkPolicy objects.
var NetworkPolicySelector = func(o runtime.Object) bool {
	_, ok := o.(*networkingv1.NetworkPolicy)
	return ok
}

// PersistentVolumeSelector is an ObjectSelector that selects PersistentVolume objects.
var PersistentVolumeSelector = func(o runtime.Object) bool {
	_, ok := o.(*corev1.PersistentVolume)
	return ok
}

// PersistentVolumeClaimSelector is an ObjectSelector that selects PersistentVolumeClaim objects.
var PersistentVolumeClaimSelector = func(o runtime.Object) bool {
	_, ok := o.(*corev1.PersistentVolumeClaim)
	return ok
}

// PodSelector is an ObjectSelector that selects Pod objects.
var PodSelector = func(o runtime.Object) bool {
	_, ok := o.(*corev1.Pod)
	return ok
}

// PodDisruptionBudgetSelector is an ObjectSelector that selects PodDisruptionBudget objects.
var PodDisruptionBudgetSelector = func(o runtime.Object) bool {
	_, ok := o.(*policyv1beta1.PodDisruptionBudget)
	return ok
}

// RoleSelector is an ObjectSelector that selects Role objects.
var RoleSelector = func(o runtime.Object) bool {
	_, ok := o.(*rbacv1.Role)
	return ok
}

// RoleBindingSelector is an ObjectSelector that selects RoleBinding objects.
var RoleBindingSelector = func(o runtime.Object) bool {
	_, ok := o.(*rbacv1.RoleBinding)
	return ok
}

// SecretSelector is an ObjectSelector that selects Secret objects.
var SecretSelector = func(o runtime.Object) bool {
	_, ok := o.(*corev1.Secret)
	return ok
}

// ServiceSelector is an ObjectSelector that selects Service objects.
var ServiceSelector = func(o runtime.Object) bool {
	_, ok := o.(*corev1.Service)
	return ok
}

// ServiceAccountSelector is an ObjectSelector that selects ServiceAccount objects.
var ServiceAccountSelector = func(o runtime.Object) bool {
	_, ok := o.(*corev1.ServiceAccount)
	return ok
}

// StatefulSetSelector is an ObjectSelector that selects StatefulSet objects.
var StatefulSetSelector = func(o runtime.Object) bool {
	_, ok := o.(*appsv1.StatefulSet)
	return ok
}

// NewConfigMapSelector builds a new ObjectSelector that selects ConfigMap objects objects with the provided delegate.
func NewConfigMapSelector(delegate func(*corev1.ConfigMap) bool) ObjectSelector {
	return func(o runtime.Object) bool {
		configMap, ok := (o).(*corev1.ConfigMap)
		if ok {
			return delegate(configMap)
		}
		return false
	}
}

// NewClusterRoleSelector builds a new ObjectSelector that selects ClusterRole objects with the provided delegate.
func NewClusterRoleSelector(delegate func(*rbacv1.ClusterRole) bool) ObjectSelector {
	return func(o runtime.Object) bool {
		clusterRole, ok := (o).(*rbacv1.ClusterRole)
		if ok {
			return delegate(clusterRole)
		}
		return false
	}
}

// NewClusterRoleBindingSelector builds a new ObjectSelector that selects ClusterRoleBinding objects with the provided delegate.
func NewClusterRoleBindingSelector(delegate func(*rbacv1.ClusterRoleBinding) bool) ObjectSelector {
	return func(o runtime.Object) bool {
		clusterRoleBinding, ok := (o).(*rbacv1.ClusterRoleBinding)
		if ok {
			return delegate(clusterRoleBinding)
		}
		return false
	}
}

// NewCronJobSelector builds a new ObjectSelector that selects CronJob objects with the provided delegate.
func NewCronJobSelector(delegate func(*batchv1beta1.CronJob) bool) ObjectSelector {
	return func(o runtime.Object) bool {
		cronJob, ok := (o).(*batchv1beta1.CronJob)
		if ok {
			return delegate(cronJob)
		}
		return false
	}
}

// NewDaemonSetSelector builds a new ObjectSelector that selects DaemonSet objects with the provided delegate.
func NewDaemonSetSelector(delegate func(*appsv1.DaemonSet) bool) ObjectSelector {
	return func(o runtime.Object) bool {
		daemonSet, ok := (o).(*appsv1.DaemonSet)
		if ok {
			return delegate(daemonSet)
		}
		return false
	}
}

// NewDeploymentSelector builds a new ObjectSelector that selects Deployment objects with the provided delegate.
func NewDeploymentSelector(delegate func(*appsv1.Deployment) bool) ObjectSelector {
	return func(o runtime.Object) bool {
		deployment, ok := (o).(*appsv1.Deployment)
		if ok {
			return delegate(deployment)
		}
		return false
	}
}

// NewHorizontalPodAutoscalerSelector builds a new ObjectSelector that selects HorizontalPodAutoscaler objects with the provided delegate.
func NewHorizontalPodAutoscalerSelector(delegate func(*autoscalingv2beta1.HorizontalPodAutoscaler) bool) ObjectSelector {
	return func(o runtime.Object) bool {
		hpa, ok := (o).(*autoscalingv2beta1.HorizontalPodAutoscaler)
		if ok {
			return delegate(hpa)
		}
		return false
	}
}

// NewIngressSelector builds a new ObjectSelector that selects Ingress objects with the provided delegate.
func NewIngressSelector(delegate func(*extensionsv1beta1.Ingress) bool) ObjectSelector {
	return func(o runtime.Object) bool {
		ingress, ok := (o).(*extensionsv1beta1.Ingress)
		if ok {
			return delegate(ingress)
		}
		return false
	}
}

// NewJobSelector builds a new ObjectSelector that selects Job objects with the provided delegate.
func NewJobSelector(delegate func(*batchv1.Job) bool) ObjectSelector {
	return func(o runtime.Object) bool {
		job, ok := (o).(*batchv1.Job)
		if ok {
			return delegate(job)
		}
		return false
	}
}

// NewNetworkPolicySelector builds a new ObjectSelector that selects NetworkPolicy objects with the provided delegate.
func NewNetworkPolicySelector(delegate func(*networkingv1.NetworkPolicy) bool) ObjectSelector {
	return func(o runtime.Object) bool {
		networkPolicy, ok := (o).(*networkingv1.NetworkPolicy)
		if ok {
			return delegate(networkPolicy)
		}
		return false
	}
}

// NewPersistentVolumeSelector builds a new ObjectSelector that selects PersistentVolume objects with the provided delegate.
func NewPersistentVolumeSelector(delegate func(*corev1.PersistentVolume) bool) ObjectSelector {
	return func(o runtime.Object) bool {
		persistentVolume, ok := (o).(*corev1.PersistentVolume)
		if ok {
			return delegate(persistentVolume)
		}
		return false
	}
}

// NewPersistentVolumeClaimSelector builds a new ObjectSelector that selects PersistentVolumeClaim objects with the provided delegate.
func NewPersistentVolumeClaimSelector(delegate func(*corev1.PersistentVolumeClaim) bool) ObjectSelector {
	return func(o runtime.Object) bool {
		pvc, ok := (o).(*corev1.PersistentVolumeClaim)
		if ok {
			return delegate(pvc)
		}
		return false
	}
}

// NewPodSelector builds a new ObjectSelector that selects Pod objects with the provided delegate.
func NewPodSelector(delegate func(*corev1.Pod) bool) ObjectSelector {
	return func(o runtime.Object) bool {
		pod, ok := (o).(*corev1.Pod)
		if ok {
			return delegate(pod)
		}
		return false
	}
}

// NewPodDisruptionBudgetSelector builds a new ObjectSelector that selects PodDisruptionBudget objects with the provided delegate.
func NewPodDisruptionBudgetSelector(delegate func(*policyv1beta1.PodDisruptionBudget) bool) ObjectSelector {
	return func(o runtime.Object) bool {
		pdb, ok := (o).(*policyv1beta1.PodDisruptionBudget)
		if ok {
			return delegate(pdb)
		}
		return false
	}
}

// NewRoleSelector builds a new ObjectSelector that selects Role objects with the provided delegate.
func NewRoleSelector(delegate func(*rbacv1.Role) bool) ObjectSelector {
	return func(o runtime.Object) bool {
		role, ok := (o).(*rbacv1.Role)
		if ok {
			return delegate(role)
		}
		return false
	}
}

// NewRoleBindingSelector builds a new ObjectSelector that selects RoleBinding objects with the provided delegate.
func NewRoleBindingSelector(delegate func(*rbacv1.RoleBinding) bool) ObjectSelector {
	return func(o runtime.Object) bool {
		roleBinding, ok := (o).(*rbacv1.RoleBinding)
		if ok {
			return delegate(roleBinding)
		}
		return false
	}
}

// NewSecretSelector builds a new ObjectSelector that selects Secret objects with the provided delegate.
func NewSecretSelector(delegate func(*corev1.Secret) bool) ObjectSelector {
	return func(o runtime.Object) bool {
		secret, ok := (o).(*corev1.Secret)
		if ok {
			return delegate(secret)
		}
		return false
	}
}

// NewServiceSelector builds a new ObjectSelector that selects Service objects with the provided delegate.
func NewServiceSelector(delegate func(*corev1.Service) bool) ObjectSelector {
	return func(o runtime.Object) bool {
		service, ok := (o).(*corev1.Service)
		if ok {
			return delegate(service)
		}
		return false
	}
}

// NewServiceAccountSelector builds a new ObjectSelector that selects ServiceAccount objects with the provided delegate.
func NewServiceAccountSelector(delegate func(*corev1.ServiceAccount) bool) ObjectSelector {
	return func(o runtime.Object) bool {
		serviceAccount, ok := (o).(*corev1.ServiceAccount)
		if ok {
			return delegate(serviceAccount)
		}
		return false
	}
}

// NewStatefulSetSelector builds a new ObjectSelector that selects StatefulSet objects with the provided delegate.
func NewStatefulSetSelector(delegate func(*appsv1.StatefulSet) bool) ObjectSelector {
	return func(o runtime.Object) bool {
		statefulSet, ok := (o).(*appsv1.StatefulSet)
		if ok {
			return delegate(statefulSet)
		}
		return false
	}
}

// NewConfigMapEditor builds a new ObjectEditor for ConfigMap objects.
func NewConfigMapEditor(delegate func(*corev1.ConfigMap) error) ObjectEditor {
	return func(o *runtime.Object) error {
		configMap, ok := (*o).(*corev1.ConfigMap)
		if !ok {
			return newTypeMistmatchError(corev1.ConfigMap{}, *o)
		}
		return delegate(configMap)
	}
}

// NewClusterRoleEditor builds a new ObjectEditor for ClusterRole objects with the provided delegate.
func NewClusterRoleEditor(delegate func(*rbacv1.ClusterRole) error) ObjectEditor {
	return func(o *runtime.Object) error {
		clusterRole, ok := (*o).(*rbacv1.ClusterRole)
		if !ok {
			return newTypeMistmatchError(rbacv1.ClusterRole{}, *o)
		}
		return delegate(clusterRole)
	}
}

// NewClusterRoleBindingEditor builds a new ObjectEditor for ClusterRoleBinding objects with the provided delegate.
func NewClusterRoleBindingEditor(delegate func(*rbacv1.ClusterRoleBinding) error) ObjectEditor {
	return func(o *runtime.Object) error {
		clusterRoleBinding, ok := (*o).(*rbacv1.ClusterRoleBinding)
		if !ok {
			return newTypeMistmatchError(rbacv1.ClusterRoleBinding{}, *o)
		}
		return delegate(clusterRoleBinding)
	}
}

// NewCronJobEditor builds a new ObjectEditor for CronJob objects with the provided delegate.
func NewCronJobEditor(delegate func(*batchv1beta1.CronJob) error) ObjectEditor {
	return func(o *runtime.Object) error {
		cronJob, ok := (*o).(*batchv1beta1.CronJob)
		if !ok {
			return newTypeMistmatchError(batchv1beta1.CronJob{}, *o)
		}
		return delegate(cronJob)
	}
}

// NewDaemonSetEditor builds a new ObjectEditor for DaemonSet objects with the provided delegate.
func NewDaemonSetEditor(delegate func(*appsv1.DaemonSet) error) ObjectEditor {
	return func(o *runtime.Object) error {
		daemonSet, ok := (*o).(*appsv1.DaemonSet)
		if !ok {
			return newTypeMistmatchError(appsv1.DaemonSet{}, *o)
		}
		return delegate(daemonSet)
	}
}

// NewDeploymentEditor builds a new ObjectEditor for Deployment objects with the provided delegate.
func NewDeploymentEditor(delegate func(*appsv1.Deployment) error) ObjectEditor {
	return func(o *runtime.Object) error {
		deployment, ok := (*o).(*appsv1.Deployment)
		if !ok {
			return newTypeMistmatchError(appsv1.Deployment{}, *o)
		}
		return delegate(deployment)
	}
}

// NewHorizontalPodAutoscalerEditor builds a new ObjectEditor for HorizontalPodAutoscaler objects with the provided delegate.
func NewHorizontalPodAutoscalerEditor(delegate func(*autoscalingv2beta1.HorizontalPodAutoscaler) error) ObjectEditor {
	return func(o *runtime.Object) error {
		hpa, ok := (*o).(*autoscalingv2beta1.HorizontalPodAutoscaler)
		if !ok {
			return newTypeMistmatchError(autoscalingv2beta1.HorizontalPodAutoscaler{}, *o)
		}
		return delegate(hpa)
	}
}

// NewIngressEditor builds a new ObjectEditor for Ingress objects with the provided delegate.
func NewIngressEditor(delegate func(*extensionsv1beta1.Ingress) error) ObjectEditor {
	return func(o *runtime.Object) error {
		ingress, ok := (*o).(*extensionsv1beta1.Ingress)
		if !ok {
			return newTypeMistmatchError(extensionsv1beta1.Ingress{}, *o)
		}
		return delegate(ingress)
	}
}

// NewJobEditor builds a new ObjectEditor for Job objects with the provided delegate.
func NewJobEditor(delegate func(*batchv1.Job) error) ObjectEditor {
	return func(o *runtime.Object) error {
		job, ok := (*o).(*batchv1.Job)
		if !ok {
			return newTypeMistmatchError(batchv1.Job{}, *o)
		}
		return delegate(job)
	}
}

// NewNetworkPolicyEditor builds a new ObjectEditor for NetworkPolicy objects with the provided delegate.
func NewNetworkPolicyEditor(delegate func(*networkingv1.NetworkPolicy) error) ObjectEditor {
	return func(o *runtime.Object) error {
		networkPolicy, ok := (*o).(*networkingv1.NetworkPolicy)
		if !ok {
			return newTypeMistmatchError(networkingv1.NetworkPolicy{}, *o)
		}
		return delegate(networkPolicy)
	}
}

// NewPersistentVolumeEditor builds a new ObjectEditor for PersistentVolume objects with the provided delegate.
func NewPersistentVolumeEditor(delegate func(*corev1.PersistentVolume) error) ObjectEditor {
	return func(o *runtime.Object) error {
		persistentVolume, ok := (*o).(*corev1.PersistentVolume)
		if !ok {
			return newTypeMistmatchError(corev1.PersistentVolume{}, *o)
		}
		return delegate(persistentVolume)
	}
}

// NewPersistentVolumeClaimEditor builds a new ObjectEditor for PersistentVolumeClaim objects with the provided delegate.
func NewPersistentVolumeClaimEditor(delegate func(*corev1.PersistentVolumeClaim) error) ObjectEditor {
	return func(o *runtime.Object) error {
		pvc, ok := (*o).(*corev1.PersistentVolumeClaim)
		if !ok {
			return newTypeMistmatchError(corev1.PersistentVolumeClaim{}, *o)
		}
		return delegate(pvc)
	}
}

// NewPodEditor builds a new ObjectEditor for Pod objects with the provided delegate.
func NewPodEditor(delegate func(*corev1.Pod) error) ObjectEditor {
	return func(o *runtime.Object) error {
		pod, ok := (*o).(*corev1.Pod)
		if !ok {
			return newTypeMistmatchError(corev1.Pod{}, *o)
		}
		return delegate(pod)
	}
}

// NewPodDisruptionBudgetEditor builds a new ObjectEditor for PodDisruptionBudget objects with the provided delegate.
func NewPodDisruptionBudgetEditor(delegate func(*policyv1beta1.PodDisruptionBudget) error) ObjectEditor {
	return func(o *runtime.Object) error {
		pdb, ok := (*o).(*policyv1beta1.PodDisruptionBudget)
		if !ok {
			return newTypeMistmatchError(policyv1beta1.PodDisruptionBudget{}, *o)
		}
		return delegate(pdb)
	}
}

// NewRoleEditor builds a new ObjectEditor for Role objects with the provided delegate.
func NewRoleEditor(delegate func(*rbacv1.Role) error) ObjectEditor {
	return func(o *runtime.Object) error {
		role, ok := (*o).(*rbacv1.Role)
		if !ok {
			return newTypeMistmatchError(rbacv1.Role{}, *o)
		}
		return delegate(role)
	}
}

// NewRoleBindingEditor builds a new ObjectEditor for RoleBinding objects with the provided delegate.
func NewRoleBindingEditor(delegate func(*rbacv1.RoleBinding) error) ObjectEditor {
	return func(o *runtime.Object) error {
		roleBinding, ok := (*o).(*rbacv1.RoleBinding)
		if !ok {
			return newTypeMistmatchError(rbacv1.RoleBinding{}, *o)
		}
		return delegate(roleBinding)
	}
}

// NewSecretEditor builds a new ObjectEditor for Secret objects with the provided delegate.
func NewSecretEditor(delegate func(*corev1.Secret) error) ObjectEditor {
	return func(o *runtime.Object) error {
		secret, ok := (*o).(*corev1.Secret)
		if !ok {
			return newTypeMistmatchError(corev1.Secret{}, *o)
		}
		return delegate(secret)
	}
}

// NewServiceEditor builds a new ObjectEditor for Service objects with the provided delegate.
func NewServiceEditor(delegate func(*corev1.Service) error) ObjectEditor {
	return func(o *runtime.Object) error {
		service, ok := (*o).(*corev1.Service)
		if !ok {
			return newTypeMistmatchError(corev1.Service{}, *o)
		}
		return delegate(service)
	}
}

// NewServiceAccountEditor builds a new ObjectEditor for ServiceAccount objects with the provided delegate.
func NewServiceAccountEditor(delegate func(*corev1.ServiceAccount) error) ObjectEditor {
	return func(o *runtime.Object) error {
		serviceAccount, ok := (*o).(*corev1.ServiceAccount)
		if !ok {
			return newTypeMistmatchError(corev1.ServiceAccount{}, *o)
		}
		return delegate(serviceAccount)
	}
}

// NewStatefulSetEditor builds a new ObjectEditor for StatefulSet objects with the provided delegate.
func NewStatefulSetEditor(delegate func(*appsv1.StatefulSet) error) ObjectEditor {
	return func(o *runtime.Object) error {
		statefulSet, ok := (*o).(*appsv1.StatefulSet)
		if !ok {
			return newTypeMistmatchError(appsv1.StatefulSet{}, *o)
		}
		return delegate(statefulSet)
	}
}
