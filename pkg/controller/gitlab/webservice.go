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

func getWebserviceDeployment(cr *gitlabv1beta1.Gitlab) *appsv1.Deployment {
	labels := gitlabutils.Label(cr.Name, "webservice", gitlabutils.GitlabType)

	webservice := gitlabutils.GenericDeployment(gitlabutils.Component{
		Namespace: cr.Namespace,
		Labels:    labels,
		Replicas:  1,
		InitContainers: []corev1.Container{
			{
				Name:            "certificates",
				Image:           GitLabCertificatesImage,
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
						ReadOnly:  false,
					},
				},
			},
			{
				Name:            "configure",
				Image:           BusyboxImage,
				ImagePullPolicy: corev1.PullAlways,
				Command:         []string{"sh"},
				Args: []string{
					"-c",
					"sh -x /config-webservice/configure; sh -x /config-workhorse/configure; mkdir -p -m 3770 /tmp/gitlab",
				},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						"cpu": gitlabutils.ResourceQuantity("50m"),
					},
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "webservice-config",
						MountPath: "/config-webservice",
						ReadOnly:  true,
					},
					{
						Name:      "workhorse-config",
						MountPath: "/config-workhorse",
						ReadOnly:  true,
					},
					{
						Name:      "init-webservice-secrets",
						MountPath: "/init-config",
						ReadOnly:  true,
					},
					{
						Name:      "webservice-secrets",
						MountPath: "/init-secrets",
					},
					{
						Name:      "workhorse-secrets",
						MountPath: "/init-secrets-workhorse",
					},
					{
						Name:      "shared-tmp",
						MountPath: "/tmp",
					},
				},
			},
			{
				Name:            "dependencies",
				Image:           GitLabWebServiceImage,
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
						Name:  "WORKHORSE_ARCHIVE_CACHE_DISABLED",
						Value: "1",
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
						Name:      "webservice-config",
						MountPath: "/var/opt/gitlab/templates",
					},
					{
						Name:      "webservice-secrets",
						MountPath: "/etc/gitlab",
						ReadOnly:  true,
					},
					{
						Name:      "webservice-secrets",
						MountPath: "/srv/gitlab/config/secrets.yml",
						SubPath:   "rails-secrets/secrets.yml",
						ReadOnly:  true,
					},
				},
			},
		},
		Containers: []corev1.Container{
			{
				Name:            "webservice",
				Image:           GitLabWebServiceImage,
				ImagePullPolicy: corev1.PullIfNotPresent,
				Env: []corev1.EnvVar{
					{
						Name:  "GITLAB_WEBSERVER",
						Value: "puma",
					},
					{
						Name:  "TMPDIR",
						Value: "/tmp/gitlab",
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
						Name:  "prometheus_multiproc_dir",
						Value: "/metrics",
					},
					{
						Name:  "ENABLE_BOOTSNAP",
						Value: "1",
					},
					{
						Name:  "WORKER_PROCESSES",
						Value: "2",
					},
					{
						Name:  "WORKER_TIMEOUT",
						Value: "60",
					},
					{
						Name:  "INTERNAL_PORT",
						Value: "8080",
					},
					{
						Name:  "PUMA_THREADS_MIN",
						Value: "4",
					},
					{
						Name:  "PUMA_THREADS_MAX",
						Value: "4",
					},
					{
						Name:  "PUMA_WORKER_MAX_MEMORY",
						Value: "1024",
					},
					{
						Name:  "DISABLE_PUMA_WORKER_KILLER",
						Value: "false",
					},
				},
				Ports: []corev1.ContainerPort{
					{
						Name:          "webservice",
						Protocol:      corev1.ProtocolTCP,
						ContainerPort: 8080,
					},
				},
				Lifecycle: &corev1.Lifecycle{
					PreStop: &corev1.Handler{
						Exec: &corev1.ExecAction{
							Command: []string{"/bin/bash", "-c", "pkill -SIGINT -o ruby"},
						},
					},
				},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						"cpu":    gitlabutils.ResourceQuantity("300m"),
						"memory": gitlabutils.ResourceQuantity("1500M"),
					},
				},
				LivenessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						HTTPGet: &corev1.HTTPGetAction{
							Path: "/-/liveness",
							Port: intstr.IntOrString{
								IntVal: 8080,
							},
							Scheme: corev1.URISchemeHTTP,
						},
					},
					InitialDelaySeconds: 20,
					PeriodSeconds:       60,
					SuccessThreshold:    1,
					TimeoutSeconds:      30,
					FailureThreshold:    3,
				},
				ReadinessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						HTTPGet: &corev1.HTTPGetAction{
							Path: "/-/readiness",
							Port: intstr.IntOrString{
								IntVal: 8080,
							},
							Scheme: corev1.URISchemeHTTP,
						},
					},
					InitialDelaySeconds: 0,
					PeriodSeconds:       10,
					SuccessThreshold:    1,
					TimeoutSeconds:      2,
					FailureThreshold:    3,
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "webservice-metrics",
						MountPath: "/metrics",
					},
					{
						Name:      "webservice-config",
						MountPath: "/var/opt/gitlab/templates",
					},
					{
						Name:      "webservice-secrets",
						MountPath: "/etc/gitlab",
						ReadOnly:  true,
					},
					{
						Name:      "webservice-secrets",
						MountPath: "/srv/gitlab/config/secrets.yml",
						SubPath:   "rails-secrets/secrets.yml",
					},
					{
						Name:      "webservice-config",
						MountPath: "/srv/gitlab/config/initializers/smtp_settings.rb",
						SubPath:   "smtp_settings.rb",
					},
					{
						Name:      "webservice-config",
						MountPath: "/srv/gitlab/INSTALLATION_TYPE",
						SubPath:   "installation_type",
					},
					{
						Name:      "shared-upload-directory",
						MountPath: "/srv/gitlab/public/uploads/tmp",
						ReadOnly:  false,
					},
					{
						Name:      "shared-artifact-directory",
						MountPath: "/srv/gitlab/shared",
						ReadOnly:  false,
					},
					{
						Name:      "shared-tmp",
						MountPath: "/tmp",
						ReadOnly:  false,
					},
					{
						Name:      "etc-ssl-certs",
						MountPath: "/etc/ssl/certs/",
						ReadOnly:  true,
					},
				},
			},
			{
				Name:            "workhorse",
				Image:           GitLabWorkhorseImage,
				ImagePullPolicy: corev1.PullIfNotPresent,
				Env: []corev1.EnvVar{
					{
						Name:  "TMPDIR",
						Value: "/tmp/gitlab",
					},
					{
						Name:  "GITLAB_WORKHORSE_EXTRA_ARGS",
						Value: "",
					},
					{
						Name:  "GITLAB_WORKHORSE_LISTEN_PORT",
						Value: "8181",
					},
					{
						Name:  "CONFIG_TEMPLATE_DIRECTORY",
						Value: "/var/opt/gitlab/templates",
					},
					{
						Name:  "CONFIG_DIRECTORY",
						Value: "/srv/gitlab/config",
					},
				},
				Ports: []corev1.ContainerPort{
					{
						Name:          "workhorse",
						ContainerPort: 8181,
						Protocol:      corev1.ProtocolTCP,
					},
				},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						"cpu":    gitlabutils.ResourceQuantity("100m"),
						"memory": gitlabutils.ResourceQuantity("100M"),
					},
				},
				LivenessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						Exec: &corev1.ExecAction{
							Command: []string{"/scripts/healthcheck"},
						},
					},
					FailureThreshold:    3,
					InitialDelaySeconds: 20,
					PeriodSeconds:       60,
					SuccessThreshold:    1,
					TimeoutSeconds:      30,
				},
				ReadinessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						Exec: &corev1.ExecAction{
							Command: []string{"/scripts/healthcheck"},
						},
					},
					InitialDelaySeconds: 0,
					FailureThreshold:    3,
					PeriodSeconds:       10,
					SuccessThreshold:    1,
					TimeoutSeconds:      2,
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "workhorse-config",
						MountPath: "/var/opt/gitlab/templates",
					},
					{
						Name:      "workhorse-secrets",
						MountPath: "/etc/gitlab",
						ReadOnly:  true,
					},
					{
						Name:      "shared-upload-directory",
						MountPath: "/srv/gitlab/public/uploads/tmp",
						ReadOnly:  false,
					},
					{
						Name:      "shared-artifact-directory",
						MountPath: "/srv/gitlab/shared",
						ReadOnly:  false,
					},
					{
						Name:      "shared-tmp",
						MountPath: "/tmp",
						ReadOnly:  false,
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
				Name: "shared-tmp",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			},
			{
				Name: "webservice-metrics",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{
						Medium: corev1.StorageMediumMemory,
					},
				},
			},
			{
				Name: "webservice-config",
				VolumeSource: corev1.VolumeSource{
					Projected: &corev1.ProjectedVolumeSource{
						DefaultMode: &gitlabutils.ProjectedVolumeDefaultMode,
						Sources: []corev1.VolumeProjection{
							{
								ConfigMap: &corev1.ConfigMapProjection{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: cr.Name + "-webservice-config",
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
				Name: "workhorse-config",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: cr.Name + "-workhorse-config",
						},
						DefaultMode: &gitlabutils.ConfigMapDefaultMode,
					},
				},
			},
			{
				Name: "init-webservice-secrets",
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
										Name: cr.Name + "-shell-secret",
									},
									Items: []corev1.KeyToPath{
										{
											Key:  "secret",
											Path: "shell/.gitlab_shell_secret",
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
											Path: "redis/redis-password",
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
										Name: cr.Name + "-workhorse-secret",
									},
									Items: []corev1.KeyToPath{
										{
											Key:  "shared_secret",
											Path: "gitlab-workhorse/secret",
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
				Name: "webservice-secrets",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{
						Medium: corev1.StorageMediumMemory,
					},
				},
			},
			{
				Name: "workhorse-secrets",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{
						Medium: corev1.StorageMediumMemory,
					},
				},
			},
			{
				Name: "shared-upload-directory",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			},
			{
				Name: "shared-artifact-directory",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
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

	webservice.Spec.Template.Spec.SecurityContext = &corev1.PodSecurityContext{
		RunAsUser: &localUser,
		FSGroup:   &localUser,
	}

	webservice.Spec.Template.Spec.ServiceAccountName = "gitlab"

	return webservice
}

func (r *ReconcileGitlab) reconcileWebserviceDeployment(cr *gitlabv1beta1.Gitlab) error {
	webservice := getWebserviceDeployment(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: webservice.Name}, webservice) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, webservice, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), webservice)
}
