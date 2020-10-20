package utils

import (
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Component represents an application / micro-service
// that makes part of a larger application
type Component struct {
	// Namespace of the component
	Namespace string
	// Defines the number of pods for the
	// component to be created
	Replicas int32
	// Labels for the component
	Labels map[string]string
	// InitContainers contains a list of containers	that may
	// need to run before the main application container starts up
	InitContainers []corev1.Container
	// Containers containers a list of containers that make up a pod
	Containers []corev1.Container
	// Contains a list of volumes used by the containers
	Volumes []corev1.Volume
	// Defines volume claims to be used by a statefulset
	VolumeClaimTemplates []corev1.PersistentVolumeClaim
}

// GenericStatefulSet returns a generic k8s statefulset
func GenericStatefulSet(component Component) *appsv1.StatefulSet {
	labels := component.Labels

	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labels["app.kubernetes.io/instance"],
			Namespace: component.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName:          strings.Join([]string{labels["app.kubernetes.io/instance"], "headless"}, "-"),
			VolumeClaimTemplates: component.VolumeClaimTemplates,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					InitContainers: component.InitContainers,
					Containers:     component.Containers,
					Volumes:        component.Volumes,
				},
			},
		},
	}
}

// GenericDeployment returns a generic deployment
func GenericDeployment(component Component) *appsv1.Deployment {
	var replicas int32
	labels := component.Labels

	if component.Replicas != 0 {
		replicas = component.Replicas
	} else {
		replicas = 1
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labels["app.kubernetes.io/instance"],
			Namespace: component.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					InitContainers: component.InitContainers,
					Containers:     component.Containers,
					Volumes:        component.Volumes,
				},
			},
		},
	}
}

// GenericJob returns a Kubernetes Job
func GenericJob(component Component) *batchv1.Job {
	labels := component.Labels
	var (
		replicas          int32 = 1
		backoffLimit      int32 = 6
		activeDeadlineSec int64 = 3600
	)

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labels["app.kubernetes.io/instance"],
			Namespace: component.Namespace,
			Labels:    labels,
		},
		Spec: batchv1.JobSpec{
			Parallelism:           &replicas,
			Completions:           &replicas,
			BackoffLimit:          &backoffLimit,
			ActiveDeadlineSeconds: &activeDeadlineSec,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					InitContainers: component.InitContainers,
					Containers:     component.Containers,
					Volumes:        component.Volumes,
					RestartPolicy:  corev1.RestartPolicyOnFailure,
				},
			},
		},
	}
}

// GenericCronJob returns a kubernetes CronJob
func GenericCronJob(component Component) *batchv1beta1.CronJob {
	labels := component.Labels
	var (
		replicas     int32 = 1
		backoffLimit int32 = 3
	)

	return &batchv1beta1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labels["app.kubernetes.io/instance"],
			Namespace: component.Namespace,
			Labels:    labels,
		},
		Spec: batchv1beta1.CronJobSpec{
			ConcurrencyPolicy: batchv1beta1.ForbidConcurrent,
			JobTemplate: batchv1beta1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Parallelism:  &replicas,
					BackoffLimit: &backoffLimit,
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: component.Containers,
						},
					},
				},
			},
		},
	}
}

// ServiceAccount returns service account to be used by pods
func ServiceAccount(name, namespace string) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       name,
				"app.kubernetes.io/created-by": "gitlab-operator",
			},
		},
	}
}

// GenericSecret returns empty secret
func GenericSecret(name, namespace string, labels map[string]string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		StringData: map[string]string{},
	}
}

// GenericConfigMap returns empty configmap
func GenericConfigMap(name, namespace string, labels map[string]string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Data: map[string]string{},
	}
}
