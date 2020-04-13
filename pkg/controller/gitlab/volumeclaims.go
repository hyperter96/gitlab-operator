package gitlab

import (
	"reflect"

	gitlabv1beta1 "gitlab.com/ochienged/gitlab-operator/pkg/apis/gitlab/v1beta1"
	gitlabutils "gitlab.com/ochienged/gitlab-operator/pkg/controller/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getRegistryVolumeClaim(cr *gitlabv1beta1.Gitlab) *corev1.PersistentVolumeClaim {
	labels := gitlabutils.Label(cr.Name, "gitlab", gitlabutils.GitlabType)

	if reflect.DeepEqual(cr.Spec.Volumes.Registry, gitlabv1beta1.VolumeSpec{}) {
		return nil
	}

	volumeSize := cr.Spec.Volumes.Registry.Capacity

	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labels["app.kubernetes.io/instance"] + "-registry",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					"storage": gitlabutils.ResourceQuantity(volumeSize),
				},
			},
		},
	}
}

func getGitlabDataVolumeClaim(cr *gitlabv1beta1.Gitlab) *corev1.PersistentVolumeClaim {
	labels := gitlabutils.Label(cr.Name, "gitlab", gitlabutils.GitlabType)

	if reflect.DeepEqual(cr.Spec.Volumes.Data, gitlabv1beta1.VolumeSpec{}) {
		return nil
	}

	volumeSize := cr.Spec.Volumes.Data.Capacity

	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labels["app.kubernetes.io/instance"] + "-data",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					"storage": gitlabutils.ResourceQuantity(volumeSize),
				},
			},
		},
	}
}

func getGitlabConfigVolumeClaim(cr *gitlabv1beta1.Gitlab) *corev1.PersistentVolumeClaim {
	labels := gitlabutils.Label(cr.Name, "gitlab", gitlabutils.GitlabType)

	if reflect.DeepEqual(cr.Spec.Volumes.Configuration, gitlabv1beta1.VolumeSpec{}) {
		return nil
	}

	volumeSize := cr.Spec.Volumes.Configuration.Capacity

	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labels["app.kubernetes.io/instance"] + "-config",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					"storage": gitlabutils.ResourceQuantity(volumeSize),
				},
			},
		},
	}
}

func (r *ReconcileGitlab) reconcilePersistentVolumeClaims(cr *gitlabv1beta1.Gitlab) error {

	if !cr.Spec.Registry.Disabled && cr.Spec.Volumes.Registry.Capacity != "" {
		registryVolume := getRegistryVolumeClaim(cr)

		if err := r.createKubernetesResource(cr, registryVolume); err != nil {
			return err
		}
	}

	if cr.Spec.Volumes.Data.Capacity != "" {
		dataVolume := getGitlabDataVolumeClaim(cr)

		if err := r.createKubernetesResource(cr, dataVolume); err != nil {
			return err
		}
	}

	if cr.Spec.Volumes.Configuration.Capacity != "" {
		configVolume := getGitlabConfigVolumeClaim(cr)

		if err := r.createKubernetesResource(cr, configVolume); err != nil {
			return err
		}
	}

	return nil
}
