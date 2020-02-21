package gitlab

import (
	gitlabv1beta1 "github.com/OchiengEd/gitlab-operator/pkg/apis/gitlab/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getRegistryVolumeClaim(cr *gitlabv1beta1.Gitlab) *corev1.PersistentVolumeClaim {
	labels := getLabels(cr, "gitlab")

	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labels["app.kubernetes.io/name"] + "-registry",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.ResourceRequirements{
				Requests: getVolumeRequest(cr.Spec.Registry.Volume.Capacity),
			},
		},
	}
}

func getGitlabDataVolumeClaim(cr *gitlabv1beta1.Gitlab) *corev1.PersistentVolumeClaim {
	labels := getLabels(cr, "gitlab")

	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labels["app.kubernetes.io/name"] + "-data",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.ResourceRequirements{
				Requests: getVolumeRequest(cr.Spec.Volume.Capacity),
			},
		},
	}
}

func getGitlabConfigVolumeClaim(cr *gitlabv1beta1.Gitlab) *corev1.PersistentVolumeClaim {
	labels := getLabels(cr, "gitlab")

	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labels["app.kubernetes.io/name"] + "-config",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.ResourceRequirements{
				Requests: getVolumeRequest(cr.Spec.Configuration.Volume.Capacity),
			},
		},
	}
}
