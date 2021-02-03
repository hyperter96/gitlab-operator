package gitlab

import (
	"os"

	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
	gitlabutils "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// PostgresStatefulSet returns the Postgresql statefulset object
func PostgresStatefulSet(cr *gitlabv1beta1.GitLab) *appsv1.StatefulSet {
	labels := gitlabutils.Label(cr.Name, "postgresql", gitlabutils.GitlabType)

	var (
		adminUser    int64
		postgresUser int64 = 1001
	)

	dshmSize := gitlabutils.ResourceQuantity("1Gi")

	claims := []corev1.PersistentVolumeClaim{}
	mounts := []corev1.VolumeMount{
		{
			Name:      "custom-init-scripts",
			MountPath: "/docker-entrypoint-initdb.d/",
		},
		{
			Name:      "postgresql-password",
			MountPath: "/opt/bitnami/postgresql/secrets/",
		},
		{
			Name:      "dshm",
			MountPath: "/dev/shm",
		},
	}

	// Mount volume is specified
	if cr.Spec.Database.Volume.Capacity != "" {
		volumeSize := cr.Spec.Database.Volume.Capacity
		claims = append(claims, corev1.PersistentVolumeClaim{
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
					Requests: corev1.ResourceList{
						"storage": gitlabutils.ResourceQuantity(volumeSize),
					},
				},
			},
		})

		mounts = append(mounts, corev1.VolumeMount{
			Name:      "data",
			MountPath: "/bitnami/postgresql",
		})
	}

	psqlOptions := getPostgresOverrides(cr.Spec.Database)

	psql := gitlabutils.GenericStatefulSet(gitlabutils.Component{
		Labels:               labels,
		Namespace:            cr.Namespace,
		Replicas:             psqlOptions.Replicas,
		VolumeClaimTemplates: claims,
		InitContainers: []corev1.Container{
			{
				Name:            "init-chmod-data",
				Image:           BuildRelease(cr).MiniDebian(),
				ImagePullPolicy: corev1.PullAlways,
				Command: []string{
					"sh",
					"-c",
					"mkdir -p /bitnami/postgresql/data; chmod 700 /bitnami/postgresql/data; find /bitnami/postgresql -mindepth 0 -maxdepth 1 -not -name \".snapshot\" -not -name \"lost+found\" | xargs chown -R 1001:1001 ; chmod -R 777 /dev/shm",
				},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						"cpu":    gitlabutils.ResourceQuantity("250m"),
						"memory": gitlabutils.ResourceQuantity("256Mi"),
					},
				},
				SecurityContext: &corev1.SecurityContext{
					RunAsUser: &adminUser,
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "data",
						MountPath: "/bitnami/postgresql",
					},
					{
						Name:      "dshm",
						MountPath: "/dev/shm",
					},
				},
			},
		},
		Containers: []corev1.Container{
			{
				Name:            "postgres",
				Image:           BuildRelease(cr).Postgresql(),
				ImagePullPolicy: corev1.PullIfNotPresent,
				Env: []corev1.EnvVar{
					{
						Name:  "BITNAMI_DEBUG",
						Value: "false",
					},
					{
						Name:  "POSTGRESQL_PORT_NUMBER",
						Value: "5432",
					},
					{
						Name:  "POSTGRESQL_VOLUME_DIR",
						Value: "/bitnami/postgresql",
					},
					{
						Name:  "PGDATA",
						Value: "/bitnami/postgresql/data",
					},
					{
						Name:  "POSTGRES_POSTGRES_PASSWORD_FILE",
						Value: "/opt/bitnami/postgresql/secrets/postgresql-postgres-password",
					},
					{
						Name:  "POSTGRES_USER",
						Value: DatabaseUser,
					},
					{
						Name:  "POSTGRES_PASSWORD_FILE",
						Value: "/opt/bitnami/postgresql/secrets/postgresql-password",
					},
					{
						Name:  "POSTGRES_DB",
						Value: DatabaseName,
					},

					{
						Name:  "POSTGRESQL_ENABLE_LDAP",
						Value: "no",
					},
				},
				Ports: []corev1.ContainerPort{
					{
						Name:          "postgres",
						ContainerPort: 5432,
					},
				},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						"cpu":    gitlabutils.ResourceQuantity("250m"),
						"memory": gitlabutils.ResourceQuantity("256Mi"),
					},
				},
				LivenessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						Exec: &corev1.ExecAction{
							Command: []string{
								"pg_isready",
								"-h",
								"localhost",
								"--username",
								"gitlab",
								"--dbname",
								"gitlabhq_production",
							},
						},
					},
					FailureThreshold:    6,
					InitialDelaySeconds: 30,
					PeriodSeconds:       10,
					SuccessThreshold:    1,
					TimeoutSeconds:      5,
				},
				ReadinessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						Exec: &corev1.ExecAction{
							Command: []string{
								"pg_isready",
								"-h",
								"localhost",
								"--username",
								"gitlab",
								"--dbname",
								"gitlabhq_production",
							},
						},
					},
					FailureThreshold:    6,
					InitialDelaySeconds: 5,
					PeriodSeconds:       10,
					SuccessThreshold:    1,
					TimeoutSeconds:      5,
				},
				VolumeMounts: mounts,
			},
			{
				Name:            "metrics",
				Image:           BuildRelease(cr).PostgresqlExporter(),
				ImagePullPolicy: corev1.PullIfNotPresent,
				Env: []corev1.EnvVar{
					{
						Name:  "DATA_SOURCE_URI",
						Value: "127.0.0.1:5432/gitlabhq_production?sslmode=disable",
					},
					{
						Name:  "DATA_SOURCE_PASS_FILE",
						Value: "/opt/bitnami/postgresql/secrets/postgresql-password",
					},
					{
						Name:  "DATA_SOURCE_USER",
						Value: DatabaseUser,
					},
				},
				Ports: []corev1.ContainerPort{
					{
						Name:          "metrics",
						ContainerPort: 9187,
						Protocol:      corev1.ProtocolTCP,
					},
				},
				LivenessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						HTTPGet: &corev1.HTTPGetAction{
							Path: "/",
							Port: intstr.IntOrString{
								IntVal: 9187,
							},
							Scheme: corev1.URISchemeHTTP,
						},
					},
					InitialDelaySeconds: 5,
					PeriodSeconds:       10,
					SuccessThreshold:    1,
					TimeoutSeconds:      5,
					FailureThreshold:    6,
				},
				ReadinessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						HTTPGet: &corev1.HTTPGetAction{
							Path: "/",
							Port: intstr.IntOrString{
								IntVal: 9187,
							},
							Scheme: corev1.URISchemeHTTP,
						},
					},
					InitialDelaySeconds: 5,
					PeriodSeconds:       10,
					SuccessThreshold:    1,
					TimeoutSeconds:      5,
					FailureThreshold:    6,
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "postgresql-password",
						MountPath: "/opt/bitnami/postgresql/secrets/",
					},
				},
			},
		},
		Volumes: []corev1.Volume{
			{
				Name: "postgresql-password",
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName:  cr.Name + "-postgresql-secret",
						DefaultMode: &gitlabutils.ConfigMapDefaultMode,
					},
				},
			},
			{
				Name: "custom-init-scripts",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: cr.Name + "-postgresql-initdb-config",
						},
						DefaultMode: &gitlabutils.ConfigMapDefaultMode,
					},
				},
			},
			{
				Name: "dshm",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{
						Medium:    corev1.StorageMediumMemory,
						SizeLimit: &dshmSize,
					},
				},
			},
		},
	})

	psql.Spec.Template.Spec.ServiceAccountName = AppServiceAccount
	psql.Spec.Template.Spec.SecurityContext = &corev1.PodSecurityContext{
		RunAsUser: &postgresUser,
		FSGroup:   &postgresUser,
	}

	return psql
}

// RedisStatefulSet returns Redis statefulset object
func RedisStatefulSet(cr *gitlabv1beta1.GitLab) *appsv1.StatefulSet {
	labels := gitlabutils.Label(cr.Name, "redis", gitlabutils.GitlabType)

	var redisUser int64 = 1001

	redisEntrypoint := gitlabutils.ReadConfig(os.Getenv("GITLAB_OPERATOR_ASSETS") + "/templates/redis/entrypoint.sh")
	claims := []corev1.PersistentVolumeClaim{}
	mounts := []corev1.VolumeMount{
		// Pre-populating the mounts with the Redis config volume
		{
			Name:      "health",
			MountPath: "/health",
		},
		{
			Name:      "redis-password",
			MountPath: "/opt/bitnami/redis/secrets/",
		},
		{
			Name:      "config",
			MountPath: "/opt/bitnami/redis/mounted-etc",
		},
		{
			Name:      "redis-tmp-conf",
			MountPath: "/opt/bitnami/redis/etc/",
		},
	}

	if cr.Spec.Redis.Volume.Capacity != "" {
		volumeSize := cr.Spec.Redis.Volume.Capacity
		claims = append(claims, corev1.PersistentVolumeClaim{
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
					Requests: corev1.ResourceList{
						"storage": gitlabutils.ResourceQuantity(volumeSize),
					},
				},
			},
		})

		mounts = append(mounts, corev1.VolumeMount{
			Name:      "data",
			MountPath: "/data",
		})
	}

	redisOptions := getRedisOverrides(cr.Spec.Redis)

	redis := gitlabutils.GenericStatefulSet(gitlabutils.Component{
		Labels:               labels,
		Namespace:            cr.Namespace,
		Replicas:             redisOptions.Replicas,
		VolumeClaimTemplates: claims,
		Containers: []corev1.Container{
			{
				Name:            "redis",
				Image:           BuildRelease(cr).Redis(),
				ImagePullPolicy: corev1.PullIfNotPresent,
				Command:         []string{"/bin/bash", "-c", redisEntrypoint},
				Env: []corev1.EnvVar{
					{
						Name:  "REDIS_REPLICATION_MODE",
						Value: "master",
					},
					{
						Name:  "REDIS_PASSWORD_FILE",
						Value: "/opt/bitnami/redis/secrets/redis-password",
					},
					{
						Name:  "REDIS_PORT",
						Value: "6379",
					},
				},
				Ports: []corev1.ContainerPort{
					{
						Name:          "redis",
						ContainerPort: 6379,
						Protocol:      corev1.ProtocolTCP,
					},
				},
				VolumeMounts: mounts,
				LivenessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						Exec: &corev1.ExecAction{
							Command: []string{"sh", "-c", "/health/ping_liveness_local.sh 5"},
						},
					},
					FailureThreshold:    5,
					InitialDelaySeconds: 5,
					PeriodSeconds:       5,
					SuccessThreshold:    1,
					TimeoutSeconds:      5,
				},
				ReadinessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						Exec: &corev1.ExecAction{
							Command: []string{"sh", "-c", "/health/ping_readiness_local.sh 5"},
						},
					},
					FailureThreshold:    5,
					InitialDelaySeconds: 5,
					PeriodSeconds:       5,
					SuccessThreshold:    1,
					TimeoutSeconds:      1,
				},
			},
			{
				Name:            "metrics",
				Image:           BuildRelease(cr).RedisExporter(),
				ImagePullPolicy: corev1.PullIfNotPresent,
				Env: []corev1.EnvVar{
					{
						Name:  "REDIS_ALIAS",
						Value: cr.Name + "-redis",
					},
				},
				Ports: []corev1.ContainerPort{
					{
						Name:          "metrics",
						ContainerPort: 9121,
						Protocol:      corev1.ProtocolTCP,
					},
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "redis-password",
						MountPath: "/secrets/",
					},
				},
			},
		},
		Volumes: []corev1.Volume{
			{
				Name: "health",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						DefaultMode: &gitlabutils.ExecutableDefaultMode,
						LocalObjectReference: corev1.LocalObjectReference{
							Name: cr.Name + "-redis-health-config",
						},
					},
				},
			},
			{
				Name: "redis-password",
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						DefaultMode: &gitlabutils.ConfigMapDefaultMode,
						SecretName:  cr.Name + "-redis-secret",
						Items: []corev1.KeyToPath{
							{
								Key:  "secret",
								Path: "redis-password",
							},
						},
					},
				},
			},
			{
				Name: "config",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						DefaultMode: &gitlabutils.ConfigMapDefaultMode,
						LocalObjectReference: corev1.LocalObjectReference{
							Name: cr.Name + "-redis-config",
						},
					},
				},
			},
			{
				Name: "redis-tmp-conf",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			},
		},
	})

	redis.Spec.Template.Spec.ServiceAccountName = AppServiceAccount
	redis.Spec.Template.Spec.SecurityContext = &corev1.PodSecurityContext{
		RunAsUser: &redisUser,
		FSGroup:   &redisUser,
	}

	return redis
}

// GitalyStatefulSetDEPRECATED returns Gitaly statefulset service
// TODO 1: Remove hard corded CPU resources
func GitalyStatefulSetDEPRECATED(cr *gitlabv1beta1.GitLab) *appsv1.StatefulSet {
	labels := gitlabutils.Label(cr.Name, "gitaly", gitlabutils.GitlabType)

	var replicas int32 = 1
	var gitalyUserID int64 = 1000

	claims := []corev1.PersistentVolumeClaim{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "repositories",
				Namespace: cr.Namespace,
				Labels:    labels,
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{
					corev1.ReadWriteOnce,
				},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						"storage": gitlabutils.ResourceQuantity(cr.Spec.Volume.Capacity),
					},
				},
			},
		},
	}

	mounts := []corev1.VolumeMount{
		{
			MountPath: "/etc/ssl/certs",
			Name:      "etc-ssl-certs",
		},
		{
			MountPath: "/etc/gitaly/templates",
			Name:      "gitaly-config",
		},
		{
			MountPath: "/etc/gitlab-secrets",
			Name:      "gitaly-secrets",
			ReadOnly:  true,
		},
		{
			MountPath: "/home/git/repositories",
			Name:      "repositories",
		},
	}

	gitaly := gitlabutils.GenericStatefulSet(gitlabutils.Component{
		Labels:               labels,
		Namespace:            cr.Namespace,
		Replicas:             replicas,
		VolumeClaimTemplates: claims,
		InitContainers: []corev1.Container{
			{
				Name:            "certificates",
				Image:           BuildRelease(cr).Certificates(),
				ImagePullPolicy: corev1.PullAlways,
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
				Image:           BuildRelease(cr).Busybox(),
				ImagePullPolicy: corev1.PullAlways,
				Command:         []string{"sh", "/config/configure"},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "gitaly-config",
						MountPath: "/config",
						ReadOnly:  true,
					},
					{
						Name:      "init-gitaly-secrets",
						MountPath: "/init-config",
						ReadOnly:  true,
					},
					{
						Name:      "gitaly-secrets",
						MountPath: "/init-secrets",
					},
				},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						"cpu": gitlabutils.ResourceQuantity("50m"),
					},
				},
			},
		},
		Containers: []corev1.Container{
			{
				Name:            "gitaly",
				Image:           BuildRelease(cr).Gitaly(),
				ImagePullPolicy: corev1.PullIfNotPresent,
				Env: []corev1.EnvVar{
					{
						Name:  "CONFIG_TEMPLATE_DIRECTORY",
						Value: "/etc/gitaly/templates",
					},
					{
						Name:  "CONFIG_DIRECTORY",
						Value: "/etc/gitaly",
					},
					{
						Name:  "GITALY_CONFIG_FILE",
						Value: "/etc/gitaly/config.toml",
					},
					{
						Name:  "SSL_CERT_DIR",
						Value: "/etc/ssl/certs",
					},
					{
						Name:  "GITALY_PROMETHEUS_LISTEN_ADDR",
						Value: ":9236",
					},
					{
						Name: "POD_NAME",
						ValueFrom: &corev1.EnvVarSource{
							FieldRef: &corev1.ObjectFieldSelector{
								FieldPath: "metadata.name",
							},
						},
					},
				},
				VolumeMounts: mounts,
				Ports: []corev1.ContainerPort{
					{
						ContainerPort: 8075,
						Protocol:      corev1.ProtocolTCP,
					},
					{
						ContainerPort: 9236,
						Protocol:      corev1.ProtocolTCP,
					},
				},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						"cpu":    gitlabutils.ResourceQuantity("100m"),
						"memory": gitlabutils.ResourceQuantity("200Mi"),
					},
				},
				LivenessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						Exec: &corev1.ExecAction{
							Command: []string{"/scripts/healthcheck"},
						},
					},
					FailureThreshold:    3,
					InitialDelaySeconds: 30,
					PeriodSeconds:       10,
					SuccessThreshold:    1,
					TimeoutSeconds:      3,
				},
				ReadinessProbe: &corev1.Probe{
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
			},
		},
		Volumes: []corev1.Volume{
			{
				Name: "gitaly-config",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						DefaultMode: &gitlabutils.ConfigMapDefaultMode,
						LocalObjectReference: corev1.LocalObjectReference{
							Name: cr.Name + "-gitaly-config",
						},
					},
				},
			},
			{
				Name: "gitaly-secrets",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{
						Medium: corev1.StorageMediumMemory,
					},
				},
			},
			{
				Name: "init-gitaly-secrets",
				VolumeSource: corev1.VolumeSource{
					Projected: &corev1.ProjectedVolumeSource{
						Sources: []corev1.VolumeProjection{
							{
								Secret: &corev1.SecretProjection{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: cr.Name + "-gitaly-secret",
									},
									Items: []corev1.KeyToPath{
										{
											Key:  "token",
											Path: "gitaly_token",
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
											Path: ".gitlab_shell_secret",
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
											Path: "redis_password",
										},
									},
								},
							},
						},
						DefaultMode: &gitlabutils.SecretDefaultMode,
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

	gitaly.Spec.ServiceName = labels["app.kubernetes.io/instance"]

	gitaly.Spec.Template.Spec.ServiceAccountName = AppServiceAccount
	gitaly.Spec.Template.Spec.SecurityContext = &corev1.PodSecurityContext{
		RunAsUser: &gitalyUserID,
		FSGroup:   &gitalyUserID,
	}
	gitaly.Spec.Template.Spec.Affinity = &corev1.Affinity{
		PodAntiAffinity: &corev1.PodAntiAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
				{
					Weight: 1,
					PodAffinityTerm: corev1.PodAffinityTerm{
						TopologyKey: "kubernetes.io/hostname",
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: labels,
						},
					},
				},
			},
		},
	}

	return gitaly
}
