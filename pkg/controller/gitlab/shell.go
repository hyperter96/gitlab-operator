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

var (
	fsGroup   int64 = 1000
	runAsUser int64 = 1000
)

func getShellDeployment(cr *gitlabv1beta1.Gitlab) *appsv1.Deployment {
	labels := gitlabutils.Label(cr.Name, "shell", gitlabutils.GitlabType)

	shell := gitlabutils.GenericDeployment(gitlabutils.Component{
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
						Name:      "shell-config",
						MountPath: "/config",
						ReadOnly:  true,
					},
					{
						Name:      "shell-init-secrets",
						MountPath: "/init-config",
						ReadOnly:  true,
					},
					{
						Name:      "shell-secrets",
						MountPath: "/init-secrets",
					},
				},
			},
		},
		Containers: []corev1.Container{
			{
				Name:            "gitlab-shell",
				Image:           gitlabutils.GitLabShellImage,
				ImagePullPolicy: corev1.PullIfNotPresent,
				Env: []corev1.EnvVar{
					{
						Name:  "GITALY_FEATURE_DEFAULT_ON",
						Value: "1",
					},
					{
						Name:  "CONFIG_TEMPLATE_DIRECTORY",
						Value: "/etc/gitlab-shell",
					},
					{
						Name:  "CONFIG_DIRECTORY",
						Value: "/srv/gitlab-shell",
					},
					{
						Name:  "KEYS_DIRECTORY",
						Value: "/etc/gitlab-secrets/ssh",
					},
				},
				Ports: []corev1.ContainerPort{
					{
						Name:          "ssh",
						ContainerPort: 2222,
						Protocol:      corev1.ProtocolTCP,
					},
				},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						"cpu":    gitlabutils.ResourceQuantity("0"),
						"memory": gitlabutils.ResourceQuantity("6M"),
					},
				},
				LivenessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						Exec: &corev1.ExecAction{
							Command: []string{"/scripts/healthcheck"},
						},
					},
					FailureThreshold:    3,
					InitialDelaySeconds: 10,
					PeriodSeconds:       10,
					SuccessThreshold:    1,
					TimeoutSeconds:      3,
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "shell-config",
						MountPath: "/etc/gitlab-shell",
					},
					{
						Name:      "shell-secrets",
						MountPath: "/etc/gitlab-secrets",
						ReadOnly:  true,
					},
					{
						Name:      "sshd-config",
						MountPath: "/etc/ssh/sshd_config",
						SubPath:   "sshd_config",
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
				Name: "shell-config",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: cr.Name + "-shell-config",
						},
						DefaultMode: &gitlabutils.ConfigMapDefaultMode,
					},
				},
			},
			{
				Name: "sshd-config",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: cr.Name + "-shell-config",
						},
						Items: []corev1.KeyToPath{
							{
								Key:  "sshd_config",
								Path: "sshd_config",
							},
						},
						DefaultMode: &gitlabutils.ConfigMapDefaultMode,
					},
				},
			},
			{
				Name: "shell-init-secrets",
				VolumeSource: corev1.VolumeSource{
					Projected: &corev1.ProjectedVolumeSource{
						DefaultMode: &gitlabutils.SecretDefaultMode,
						Sources: []corev1.VolumeProjection{
							{
								Secret: &corev1.SecretProjection{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: cr.Name + "-shell-host-keys-secret",
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
				Name: "shell-secrets",
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

	shell.Spec.Template.Spec.SecurityContext = &corev1.PodSecurityContext{
		FSGroup:   &fsGroup,
		RunAsUser: &runAsUser,
	}

	return shell
}

func (r *ReconcileGitlab) reconcileShellDeployment(cr *gitlabv1beta1.Gitlab) error {
	shell := getShellDeployment(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: shell.Name}, shell) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, shell, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), shell)
}
