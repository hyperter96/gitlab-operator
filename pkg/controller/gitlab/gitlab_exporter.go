package gitlab

import (
	"context"

	gitlabv1beta1 "gitlab.com/ochienged/gitlab-operator/pkg/apis/gitlab/v1beta1"
	gitlabutils "gitlab.com/ochienged/gitlab-operator/pkg/controller/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func getGitlabExporterDeployment(cr *gitlabv1beta1.Gitlab) *appsv1.Deployment {
	labels := gitlabutils.Label(cr.Name, "gitlab-exporter", gitlabutils.GitlabType)

	return gitlabutils.GenericDeployment(gitlabutils.Component{
		Namespace: cr.Namespace,
		Labels:    labels,
		Replicas:  1,
		InitContainers: []corev1.Container{
			{
				Name:            "certificates",
				Image:           gitlabutils.GitLabCertificatesImage,
				ImagePullPolicy: corev1.PullIfNotPresent,
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						"cpu": gitlabutils.ResourceQuantity("50m"),
					},
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "etc-ssl-certs",
						MountPath: "/etc/ssl/certs",
					},
				},
			},
			{
				Name:            "configure",
				Image:           gitlabutils.BusyboxImage,
				ImagePullPolicy: corev1.PullAlways,
				Command:         []string{"sh", "/config/configure"},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						"cpu": gitlabutils.ResourceQuantity("50m"),
					},
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "gitlab-exporter-config",
						MountPath: "/config",
						ReadOnly:  true,
					},
					{
						Name:      "init-gitlab-exporter-secrets",
						MountPath: "/init-config",
						ReadOnly:  true,
					},
					{
						Name:      "gitlab-exporter-secrets",
						MountPath: "/init-secrets",
					},
				},
			},
		},
		Containers: []corev1.Container{
			{
				Name:            "task-runner",
				Image:           gitlabutils.GitLabExporterImage,
				ImagePullPolicy: corev1.PullIfNotPresent,
				Env: []corev1.EnvVar{
					{
						Name:  "CONFIG_TEMPLATE_DIRECTORY",
						Value: "/var/opt/gitlab-exporter/templates",
					},
					{
						Name:  "CONFIG_DIRECTORY",
						Value: "/etc/gitlab-exporter",
					},
				},
				Lifecycle: &corev1.Lifecycle{
					PreStop: &corev1.Handler{
						Exec: &corev1.ExecAction{
							Command: []string{"/bin/bash", "-c", "pkill -f 'gitlab-exporter'"},
						},
					},
				},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						"cpu":    gitlabutils.ResourceQuantity("75m"),
						"memory": gitlabutils.ResourceQuantity("100M"),
					},
				},
				LivenessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						Exec: &corev1.ExecAction{
							Command: []string{
								"pgrep",
								"-f",
								"gitlab-exporter",
							},
						},
					},
					FailureThreshold: 3,
					PeriodSeconds:    10,
					SuccessThreshold: 1,
					TimeoutSeconds:   1,
				},
				ReadinessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						Exec: &corev1.ExecAction{
							Command: []string{
								"pgrep",
								"-f",
								"gitlab-exporter",
							},
						},
					},
					FailureThreshold: 3,
					PeriodSeconds:    10,
					SuccessThreshold: 1,
					TimeoutSeconds:   1,
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "gitlab-exporter-config",
						MountPath: "/var/opt/gitlab-exporter/templates/gitlab-exporter.yml.erb",
						SubPath:   "gitlab-exporter.yml.erb",
					},
					{
						Name:      "gitlab-exporter-secrets",
						MountPath: "/etc/gitlab",
						ReadOnly:  true,
					},
					{
						Name:      "etc-ssl-certs",
						MountPath: "/etc/ssl/certs/",
						ReadOnly:  true,
					},
				},
			},
		},
		Volumes: []corev1.Volume{
			{
				Name: "gitlab-exporter-config",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						DefaultMode: &gitlabutils.ConfigMapDefaultMode,
						LocalObjectReference: corev1.LocalObjectReference{
							Name: cr.Name + "-gitlab-exporter-config",
						},
					},
				},
			},
			{
				Name: "init-gitlab-exporter-secrets",
				VolumeSource: corev1.VolumeSource{
					Projected: &corev1.ProjectedVolumeSource{
						DefaultMode: &gitlabutils.ProjectedVolumeDefaultMode,
						Sources: []corev1.VolumeProjection{
							{
								Secret: &corev1.SecretProjection{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: cr.Name + "-gitlab-secrets",
									},
									Items: []corev1.KeyToPath{
										{
											Key:  "postgres_password",
											Path: "postgres/psql-password",
										},
									},
								},
							},
							{
								Secret: &corev1.SecretProjection{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: cr.Name + "-gitlab-secrets",
									},
									Items: []corev1.KeyToPath{
										{
											Key:  "redis_password",
											Path: "redis/password",
										},
									},
								},
							},
						},
					},
				},
			},
			{
				Name: "gitlab-exporter-secrets",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{
						Medium: corev1.StorageMediumMemory,
					},
				},
			},
			{
				Name: "etc-ssl-certs",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{
						Medium: corev1.StorageMediumMemory,
					},
				},
			},
		},
	})
}

func (r *ReconcileGitlab) reconcileGitlabExporterDeployment(cr *gitlabv1beta1.Gitlab) error {
	exporter := getGitlabExporterDeployment(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: exporter.Name}, exporter) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, exporter, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), exporter)
}
