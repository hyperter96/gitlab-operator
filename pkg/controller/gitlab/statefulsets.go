package gitlab

import (
	gitlabv1beta1 "github.com/OchiengEd/gitlab-operator/pkg/apis/gitlab/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getGenericStatefulSet(cr *gitlabv1beta1.Gitlab, component Component) *appsv1.StatefulSet {
	labels := component.Labels

	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labels["app.kubernetes.io/name"],
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName:          labels["app.kubernetes.io/name"],
			VolumeClaimTemplates: component.VolumeClaimTemplates,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: component.Containers,
					Volumes:    component.Volumes,
				},
			},
		},
	}
}

func getPostgresStatefulSet(cr *gitlabv1beta1.Gitlab) *appsv1.StatefulSet {
	labels := getLabels(cr, "database")

	return getGenericStatefulSet(cr, Component{
		Labels:   labels,
		Replicas: cr.Spec.Database.Replicas,
		VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "data",
					Namespace: cr.Namespace,
					Labels:    labels,
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteOnce,
					},
					Resources: corev1.ResourceRequirements{
						Requests: getVolumeRequest(cr.Spec.Database.Volume.Capacity),
					},
				},
			},
		},
		Containers: []corev1.Container{
			{
				Name:            "postgres",
				Image:           "postgres:9.6.1",
				ImagePullPolicy: corev1.PullIfNotPresent,
				Env: []corev1.EnvVar{
					{
						Name: "POSTGRES_USER",
						ValueFrom: &corev1.EnvVarSource{
							ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: cr.Name + "-gitlab-config",
								},
								Key: "postgres_user",
							},
						},
					},
					{
						Name: "POSTGRES_PASSWORD",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: cr.Name + "-gitlab-secrets",
								},
								Key: "postgres_password",
							},
						},
					},
					{
						Name: "POSTGRES_DB",
						ValueFrom: &corev1.EnvVarSource{
							ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: cr.Name + "-gitlab-config",
								},
								Key: "postgres_db",
							},
						},
					},
					{
						Name:  "DB_EXTENSION",
						Value: "pg_trgm",
					},
					{
						Name:  "PGDATA",
						Value: "/var/lib/postgresql/data/pgdata",
					},
				},
				Ports: []corev1.ContainerPort{
					{
						Name:          "postgres",
						ContainerPort: 5432,
					},
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "data",
						MountPath: "/var/lib/postgresql/data",
						SubPath:   "postgres",
					},
					{
						Name:      "initdb",
						MountPath: "/docker-entrypoint-initdb.d",
						ReadOnly:  true,
					},
				},
				LivenessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						Exec: &corev1.ExecAction{
							Command: []string{
								"pg_isready",
								"-h",
								"localhost",
								"-U",
								"postgres",
							},
						},
					},
					InitialDelaySeconds: 30,
					TimeoutSeconds:      5,
				},
				ReadinessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						Exec: &corev1.ExecAction{
							Command: []string{
								"pg_isready",
								"-h",
								"localhost",
								"-U",
								"postgres",
							},
						},
					},
					InitialDelaySeconds: 5,
					TimeoutSeconds:      1,
				},
			},
		},
		Volumes: []corev1.Volume{
			{
				Name: "initdb",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: cr.Name + "-postgres-initdb",
						},
					},
				},
			},
		},
	})
}

func getRedisStatefulSet(cr *gitlabv1beta1.Gitlab) *appsv1.StatefulSet {
	labels := getLabels(cr, "redis")

	return getGenericStatefulSet(cr, Component{
		Labels:   labels,
		Replicas: cr.Spec.Redis.Replicas,
		VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "data",
					Namespace: cr.Namespace,
					Labels:    labels,
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteOnce,
					},
					Resources: corev1.ResourceRequirements{
						Requests: getVolumeRequest(cr.Spec.Redis.Volume.Capacity),
					},
				},
			},
		},
		Containers: []corev1.Container{
			{
				Name:            "redis",
				Image:           "redis:3.2.4",
				ImagePullPolicy: corev1.PullIfNotPresent,
				Command:         []string{"redis-server", "/etc/redis/redis.conf"},
				Ports: []corev1.ContainerPort{
					{
						Name:          "redis",
						ContainerPort: 6379,
					},
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "data",
						MountPath: "/var/lib/redis",
						SubPath:   "redis",
					},
					{
						Name:      "conf",
						MountPath: "/etc/redis/redis.conf",
						SubPath:   "redis.conf",
					},
				},
				LivenessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						Exec: &corev1.ExecAction{
							Command: []string{"redis-cli", "ping"},
						},
					},
					InitialDelaySeconds: 30,
					TimeoutSeconds:      5,
				},
				ReadinessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						Exec: &corev1.ExecAction{
							Command: []string{"redis-cli", "ping"},
						},
					},
					InitialDelaySeconds: 5,
					TimeoutSeconds:      1,
				},
			},
		},
		Volumes: []corev1.Volume{
			{
				Name: "conf",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: cr.Name + "-gitlab-redis",
						},
					},
				},
			},
		},
	})
}
