package gitlab

import (
	"context"

	miniov1beta1 "github.com/minio/minio-operator/pkg/apis/miniocontroller/v1beta1"
	gitlabv1beta1 "gitlab.com/ochienged/gitlab-operator/pkg/apis/gitlab/v1beta1"
	gitlabutils "gitlab.com/ochienged/gitlab-operator/pkg/controller/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func getMinioInstance(cr *gitlabv1beta1.Gitlab) *miniov1beta1.MinIOInstance {
	labels := gitlabutils.Label(cr.Name, "minio", gitlabutils.GitlabType)

	minioOptions := getMinioOverrides(cr.Spec.Minio)

	minio := &miniov1beta1.MinIOInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-minio",
			Labels:    labels,
			Namespace: cr.Namespace,
		},
		Spec: miniov1beta1.MinIOInstanceSpec{
			Metadata: &metav1.ObjectMeta{
				Labels: labels,
			},
			Replicas: minioOptions.Replicas,
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					"cpu":    gitlabutils.ResourceQuantity("250m"),
					"memory": gitlabutils.ResourceQuantity("512Mi"),
				},
			},
			Image: gitlabutils.MinioImage,
			CredsSecret: &corev1.LocalObjectReference{
				Name: cr.Name + "-minio-secret",
			},
			RequestAutoCert: false,
			Env: []corev1.EnvVar{
				{
					Name:  "MINIO_BROWSER",
					Value: "on",
				},
			},
			Mountpath: "/export",
			Liveness: &corev1.Probe{
				Handler: corev1.Handler{
					HTTPGet: &corev1.HTTPGetAction{
						Path: "/minio/health/live",
						Port: intstr.IntOrString{
							IntVal: 9000,
						},
					},
				},
				InitialDelaySeconds: 120,
				PeriodSeconds:       20,
			},
			Readiness: &corev1.Probe{
				Handler: corev1.Handler{
					HTTPGet: &corev1.HTTPGetAction{
						Path: "/minio/health/ready",
						Port: intstr.IntOrString{
							IntVal: 9000,
						},
					},
				},
				InitialDelaySeconds: 120,
				PeriodSeconds:       20,
			},
			PodManagementPolicy: appsv1.ParallelPodManagement,
			VolumeClaimTemplate: &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "data",
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteOnce,
					},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							"storage": gitlabutils.ResourceQuantity(minioOptions.Volume.Capacity),
						},
					},
				},
			},
		},
	}

	if minioOptions.Volume.StorageClass != "" {
		minio.Spec.VolumeClaimTemplate.Spec.StorageClassName = &minioOptions.Volume.StorageClass
	}

	return minio
}

func (r *ReconcileGitlab) reconcileMinioInstance(cr *gitlabv1beta1.Gitlab) error {
	minio := getMinioInstance(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: minio.Name}, minio) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, minio, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), minio)
}
