package gitlab

import (
	"context"

	gitlabv1beta1 "gitlab.com/ochienged/gitlab-operator/pkg/apis/gitlab/v1beta1"
	gitlabutils "gitlab.com/ochienged/gitlab-operator/pkg/controller/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func getSidekiqDeployment(cr *gitlabv1beta1.Gitlab) *appsv1.Deployment {
	labels := gitlabutils.Label(cr.Name, "sidekiq", gitlabutils.GitlabType)

	sidekiq := gitlabutils.GenericDeployment(gitlabutils.Component{
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
						Name:      "sidekiq-config",
						MountPath: "/config",
						ReadOnly:  true,
					},
					{
						Name:      "init-sidekiq-secrets",
						MountPath: "/init-config",
						ReadOnly:  true,
					},
					{
						Name:      "sidekiq-secrets",
						MountPath: "/init-secrets",
					},
				},
			},
			{
				Name:            "dependencies",
				Image:           gitlabutils.GitLabSidekigImage,
				ImagePullPolicy: corev1.PullIfNotPresent,
				Args:            []string{"/scripts/wait-for-deps"},
				Env: []corev1.EnvVar{
					{
						Name:  "GITALY_FEATURE_DEFAULT_ON",
						Value: "1",
					},
					{
						Name:  "CONFIG_TEMPLATE_DIRECTORY",
						Value: "/var/opt/gitlab/templates",
					},
					{
						Name:  "CONFIG_DIRECTORY",
						Value: "/srv/gitlab/config",
					},
					{
						Name:  "SIDEKIQ_CONCURRENCY",
						Value: "25",
					},
					{
						Name:  "SIDEKIQ_TIMEOUT",
						Value: "5",
					},
					{
						Name:  "ENABLE_BOOTSNAP",
						Value: "1",
					},
				},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						"cpu": gitlabutils.ResourceQuantity("50m"),
					},
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "sidekiq-config",
						MountPath: "/var/opt/gitlab/templates",
						ReadOnly:  true,
					},
					{
						Name:      "sidekiq-secrets",
						MountPath: "/etc/gitlab",
						ReadOnly:  true,
					},
					{
						Name:      "sidekiq-secrets",
						MountPath: "/srv/gitlab/config/secrets.yml",
						SubPath:   "rails-secrets/secrets.yml",
						ReadOnly:  true,
					},
				},
			},
		},
		Containers: []corev1.Container{
			{
				Name:            "sidekiq",
				Image:           gitlabutils.GitLabSidekigImage,
				ImagePullPolicy: corev1.PullIfNotPresent,
				Env: []corev1.EnvVar{
					{
						Name:  "prometheus_multiproc_dir",
						Value: "/metrics",
					},
					{
						Name:  "GITALY_FEATURE_DEFAULT_ON",
						Value: "1",
					},
					{
						Name:  "CONFIG_TEMPLATE_DIRECTORY",
						Value: "/var/opt/gitlab/templates",
					},
					{
						Name:  "CONFIG_DIRECTORY",
						Value: "/srv/gitlab/config",
					},
					{
						Name:  "SIDEKIQ_CONCURRENCY",
						Value: "25",
					},
					{
						Name:  "SIDEKIQ_TIMEOUT",
						Value: "5",
					},
					{
						Name:  "SIDEKIQ_DAEMON_MEMORY_KILLER",
						Value: "0",
					},
					{
						Name:  "SIDEKIQ_MEMORY_KILLER_CHECK_INTERVAL",
						Value: "3",
					},
					{
						Name:  "SIDEKIQ_MEMORY_KILLER_MAX_RSS",
						Value: "2000000",
					},
					{
						Name:  "SIDEKIQ_MEMORY_KILLER_GRACE_TIME",
						Value: "900",
					},
					{
						Name:  "SIDEKIQ_MEMORY_KILLER_SHUTDOWN_WAIT",
						Value: "30",
					},
					{
						Name:  "ENABLE_BOOTSNAP",
						Value: "1",
					},
				},
				Ports: []corev1.ContainerPort{
					{
						Name:          "metrics",
						ContainerPort: 3807,
						Protocol:      corev1.ProtocolTCP,
					},
				},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						"cpu":    gitlabutils.ResourceQuantity("50m"),
						"memory": gitlabutils.ResourceQuantity("650M"),
					},
				},
				Lifecycle: &corev1.Lifecycle{
					PreStop: &corev1.Handler{
						Exec: &corev1.ExecAction{
							Command: []string{"/bin/bash", "-c", "pkill -f 'sidekiq'"},
						},
					},
				},
				LivenessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						HTTPGet: &corev1.HTTPGetAction{
							Path: "/liveness",
							Port: intstr.IntOrString{
								IntVal: 3807,
							},
							Scheme: corev1.URISchemeHTTP,
						},
					},
					InitialDelaySeconds: 20,
					PeriodSeconds:       60,
					SuccessThreshold:    1,
					TimeoutSeconds:      30,
				},
				ReadinessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						HTTPGet: &corev1.HTTPGetAction{
							Path: "/readiness",
							Port: intstr.IntOrString{
								IntVal: 3807,
							},
							Scheme: corev1.URISchemeHTTP,
						},
					},
					PeriodSeconds:    10,
					SuccessThreshold: 1,
					TimeoutSeconds:   2,
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "sidekiq-metrics",
						MountPath: "/metrics",
					},
					{
						Name:      "sidekiq-config",
						MountPath: "/var/opt/gitlab/templates",
						ReadOnly:  true,
					},
					{
						Name:      "sidekiq-secrets",
						MountPath: "/etc/gitlab",
						ReadOnly:  true,
					},
					{
						Name:      "sidekiq-secrets",
						MountPath: "/srv/gitlab/config/secrets.yml",
						SubPath:   "rails-secrets/secrets.yml",
					},
					{
						Name:      "sidekiq-config",
						MountPath: "/srv/gitlab/config/initializers/smtp_settings.rb",
						SubPath:   "smtp_settings.rb",
					},
					{
						Name:      "sidekiq-config",
						MountPath: "/srv/gitlab/INSTALLATION_TYPE",
						SubPath:   "installation_type",
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
				Name: "sidekiq-metrics",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{
						Medium: corev1.StorageMediumMemory,
					},
				},
			},
			{
				Name: "sidekiq-config",
				VolumeSource: corev1.VolumeSource{
					Projected: &corev1.ProjectedVolumeSource{
						DefaultMode: &gitlabutils.ProjectedVolumeDefaultMode,
						Sources: []corev1.VolumeProjection{
							{
								ConfigMap: &corev1.ConfigMapProjection{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: cr.Name + "-sidekiq-config",
									},
								},
							},
							{
								Secret: &corev1.SecretProjection{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: cr.Name + "-smtp-settings-secret",
									},
								},
							},
						},
					},
				},
			},
			{
				Name: "init-sidekiq-secrets",
				VolumeSource: corev1.VolumeSource{
					Projected: &corev1.ProjectedVolumeSource{
						DefaultMode: &gitlabutils.ProjectedVolumeDefaultMode,
						Sources: []corev1.VolumeProjection{
							{
								Secret: &corev1.SecretProjection{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: cr.Name + "-rails-secret",
									},
									Items: []corev1.KeyToPath{
										{
											Key:  "secrets.yml",
											Path: "rails-secrets/secrets.yml",
										},
									},
								},
							},
							{
								Secret: &corev1.SecretProjection{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: cr.Name + "-gitaly-secret",
									},
									Items: []corev1.KeyToPath{
										{
											Key:  "token",
											Path: "gitaly/gitaly_token",
										},
									},
								},
							},
							{
								Secret: &corev1.SecretProjection{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: cr.Name + "-redis-secret",
									},
									Items: []corev1.KeyToPath{
										{
											Key:  "secret",
											Path: "redis/password",
										},
									},
								},
							},
							{
								Secret: &corev1.SecretProjection{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: cr.Name + "-postgresql-secret",
									},
									Items: []corev1.KeyToPath{
										{
											Key:  "postgresql-password",
											Path: "postgres/psql-password",
										},
									},
								},
							},
							{
								Secret: &corev1.SecretProjection{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: cr.Name + "-registry-secret",
									},
									Items: []corev1.KeyToPath{
										{
											Key:  "registry-auth.key",
											Path: "registry/gitlab-registry.key",
										},
									},
								},
							},
							{
								Secret: &corev1.SecretProjection{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: cr.Name + "-minio-secret",
									},
									Items: []corev1.KeyToPath{
										{
											Key:  "accesskey",
											Path: "minio/accesskey",
										},
										{
											Key:  "secretkey",
											Path: "minio/secretkey",
										},
									},
								},
							},
						},
					},
				},
			},
			{
				Name: "sidekiq-secrets",
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

	sidekiq.Spec.Template.Spec.SecurityContext = &corev1.PodSecurityContext{
		RunAsUser: &runAsUser,
		FSGroup:   &fsGroup,
	}

	sidekiq.Spec.Template.Spec.ServiceAccountName = "gitlab"

	return sidekiq
}

func (r *ReconcileGitlab) reconcileSidekiqDeployment(cr *gitlabv1beta1.Gitlab) error {
	sidekiq := getSidekiqDeployment(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: sidekiq.Name}, sidekiq) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, sidekiq, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), sidekiq)
}
