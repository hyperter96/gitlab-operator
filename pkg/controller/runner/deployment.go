package runner

import (
	gitlabv1beta1 "github.com/OchiengEd/gitlab-operator/pkg/apis/gitlab/v1beta1"
	gitlab "github.com/OchiengEd/gitlab-operator/pkg/controller/gitlab"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func getRunnerDeployment(cr *gitlabv1beta1.Runner) *appsv1.Deployment {
	labels := getLabels(cr, "runner")

	return gitlab.GenericDeployment(gitlab.Component{
		Labels:    labels,
		Namespace: cr.Namespace,
		Containers: []corev1.Container{
			{
				Name:    "runner",
				Image:   RunnerImage,
				Command: []string{"/bin/bash", "/scripts/entrypoint"},
				Lifecycle: &corev1.Lifecycle{
					PreStop: &corev1.Handler{
						Exec: &corev1.ExecAction{
							Command: []string{"gitlab-runner", "unregister", "--all-runners"},
						},
					},
				},
				Env: []corev1.EnvVar{
					{
						Name: "CI_SERVER_URL",
						ValueFrom: &corev1.EnvVarSource{
							ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: cr.Name + "-runner-scripts",
								},
								Key: "ci_server_url",
							},
						},
					},
					{
						Name: "CI_SERVER_TOKEN",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: cr.Name + "-runner-secrets",
								},
								Key: "runner-token",
							},
						},
					},
					{
						Name: "REGISTRATION_TOKEN",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: cr.Name + "-runner-secrets",
								},
								Key: "runner-registration-token",
							},
						},
					},
					{
						Name:  "KUBERNETES_NAMESPACE",
						Value: cr.Namespace,
					},
					{
						Name:  "KUBERNETES_PRIVILEGED",
						Value: "true",
					},
					{
						Name:  "KUBERNETES_IMAGE",
						Value: "ubuntu:16.04",
					},
					{
						Name:  "KUBERNETES_CPU_LIMIT",
						Value: "",
					},
					{
						Name:  "KUBERNETES_MEMORY_LIMIT",
						Value: "",
					},
					{
						Name:  "KUBERNETES_CPU_REQUEST",
						Value: "",
					},
					{
						Name:  "KUBERNETES_MEMORY_REQUEST",
						Value: "",
					},
					{
						Name:  "KUBERNETES_SERVICE_CPU_LIMIT",
						Value: "",
					},
					{
						Name:  "KUBERNETES_SERVICE_MEMORY_LIMIT",
						Value: "",
					},
					{
						Name:  "KUBERNETES_SERVICE_CPU_REQUEST",
						Value: "",
					},
					{
						Name:  "KUBERNETES_SERVICE_MEMORY_REQUEST",
						Value: "",
					},
					{
						Name:  "KUBERNETES_HELPERS_CPU_LIMIT",
						Value: "",
					},
					{
						Name:  "KUBERNETES_HELPERS_MEMORY_LIMIT",
						Value: "",
					},
					{
						Name:  "KUBERNETES_HELPERS_CPU_REQUEST",
						Value: "",
					},
					{
						Name:  "KUBERNETES_HELPERS_MEMORY_REQUEST",
						Value: "",
					},
				},
				LivenessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						Exec: &corev1.ExecAction{
							Command: []string{"/usr/bin/pgrep", "gitlab.*runner"},
						},
					},
					InitialDelaySeconds: 60,
					TimeoutSeconds:      1,
					PeriodSeconds:       3,
					SuccessThreshold:    1,
					FailureThreshold:    3,
				},
				ReadinessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						Exec: &corev1.ExecAction{
							Command: []string{"/usr/bin/pgrep", "gitlab.*runner"},
						},
					},
					InitialDelaySeconds: 10,
					TimeoutSeconds:      1,
					PeriodSeconds:       10,
					SuccessThreshold:    1,
					FailureThreshold:    3,
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "scripts",
						MountPath: "/scripts",
					},
				},
			},
		},
		Volumes: []corev1.Volume{
			{
				Name: "scripts",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: cr.Name + "-runner-scripts",
						},
					},
				},
			},
		},
	})
}

func getLabels(cr *gitlabv1beta1.Runner, component string) map[string]string {

	return map[string]string{
		"app.kubernetes.io/name":       cr.Name + "-" + component,
		"app.kubernetes.io/component":  component,
		"app.kubernetes.io/part-of":    "gitlab",
		"app.kubernetes.io/managed-by": "gitlab-operator",
	}
}
