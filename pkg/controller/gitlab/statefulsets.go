package gitlab

import (
	"context"

	gitlabv1beta1 "gitlab.com/ochienged/gitlab-operator/pkg/apis/gitlab/v1beta1"
	gitlabutils "gitlab.com/ochienged/gitlab-operator/pkg/controller/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func getPostgresStatefulSet(cr *gitlabv1beta1.Gitlab) *appsv1.StatefulSet {
	labels := gitlabutils.Label(cr.Name, "database", gitlabutils.GitlabType)

	var (
		runAsUser   int64 = 0
		pgRunAsUser int64 = 1001
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
	if cr.Spec.Volumes.Postgres.Capacity != "" {
		volumeSize := cr.Spec.Volumes.Postgres.Capacity
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
						StorageResourceName: gitlabutils.ResourceQuantity(volumeSize),
					},
				},
			},
		})

		mounts = append(mounts, corev1.VolumeMount{
			Name:      "data",
			MountPath: "/bitnami/postgresql",
		})
	}

	return gitlabutils.GenericStatefulSet(gitlabutils.Component{
		Labels:               labels,
		Namespace:            cr.Namespace,
		Replicas:             cr.Spec.Database.Replicas,
		VolumeClaimTemplates: claims,
		InitContainers: []corev1.Container{
			{
				Name:            "init-chmod-data",
				Image:           gitlabutils.MiniDebImage,
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
					RunAsUser: &runAsUser,
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
				Image:           gitlabutils.PostgresImage,
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
				SecurityContext: &corev1.SecurityContext{
					RunAsUser: &pgRunAsUser,
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
				Image:           gitlabutils.PostgresExporterImage,
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
}

func getRedisStatefulSet(cr *gitlabv1beta1.Gitlab) *appsv1.StatefulSet {
	labels := gitlabutils.Label(cr.Name, "redis", gitlabutils.GitlabType)

	claims := []corev1.PersistentVolumeClaim{}
	mounts := []corev1.VolumeMount{
		// Pre-populating the mounts with the Redis config volume
		{
			Name:      "conf",
			MountPath: "/etc/redis/redis.conf",
			SubPath:   "redis.conf",
		},
	}

	if cr.Spec.Volumes.Redis.Capacity != "" {
		volumeSize := cr.Spec.Volumes.Redis.Capacity
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
						StorageResourceName: gitlabutils.ResourceQuantity(volumeSize),
					},
				},
			},
		})

		mounts = append(mounts, corev1.VolumeMount{
			Name:      "data",
			MountPath: "/var/lib/redis",
			SubPath:   "redis",
		})
	}

	return gitlabutils.GenericStatefulSet(gitlabutils.Component{
		Labels:               labels,
		Namespace:            cr.Namespace,
		Replicas:             cr.Spec.Redis.Replicas,
		VolumeClaimTemplates: claims,
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
				VolumeMounts: mounts,
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

// TODO 1: Remove hard corded volume size
// TODO 2: Remove hard corded CPU resources
// TODO 3: Add Gitaly secrets
func getGitalyStatefulSet(cr *gitlabv1beta1.Gitlab) *appsv1.StatefulSet {
	labels := gitlabutils.Label(cr.Name, "gitaly", gitlabutils.GitlabType)

	volumeSize := "10Gi"
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
						StorageResourceName: gitlabutils.ResourceQuantity(volumeSize),
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

	return gitlabutils.GenericStatefulSet(gitlabutils.Component{
		Labels:               labels,
		Namespace:            cr.Namespace,
		Replicas:             cr.Spec.Redis.Replicas,
		VolumeClaimTemplates: claims,
		InitContainers: []corev1.Container{
			{
				Name:            "certificates",
				Image:           gitlabutils.GitLabCertificatesImage,
				ImagePullPolicy: corev1.PullAlways,
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						StorageResourceName: gitlabutils.ResourceQuantity("50m"),
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
						StorageResourceName: gitlabutils.ResourceQuantity("50m"),
					},
				},
			},
		},
		Containers: []corev1.Container{
			{
				Name:            "gitaly",
				Image:           gitlabutils.GitalyImage,
				ImagePullPolicy: corev1.PullIfNotPresent,
				VolumeMounts:    mounts,
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
				// LivenessProbe: &corev1.Probe{
				// 	Handler: corev1.Handler{
				// 		Exec: &corev1.ExecAction{
				// 			Command: []string{"/scripts/healthcheck"},
				// 		},
				// 	},
				// 	FailureThreshold:    3,
				// 	InitialDelaySeconds: 30,
				// 	PeriodSeconds:       10,
				// 	SuccessThreshold:    1,
				// 	TimeoutSeconds:      3,
				// },
				// ReadinessProbe: &corev1.Probe{
				// 	Handler: corev1.Handler{
				// 		Exec: &corev1.ExecAction{
				// 			Command: []string{"/scripts/healthcheck"},
				// 		},
				// 	},
				// 	FailureThreshold:    3,
				// 	InitialDelaySeconds: 10,
				// 	PeriodSeconds:       10,
				// 	SuccessThreshold:    1,
				// 	TimeoutSeconds:      3,
				// },
			},
		},
		Volumes: []corev1.Volume{
			{
				Name: "gitaly-config",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: cr.Name + "-gitaly-config",
						},
						// DefaultMode: 420,
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
										Name: cr.Name + "-gitlab-shell-secret",
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
}

func (r *ReconcileGitlab) reconcilePostgresStatefulSet(cr *gitlabv1beta1.Gitlab) error {
	postgres := getPostgresStatefulSet(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: postgres.Name}, postgres) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, postgres, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), postgres)
}

func (r *ReconcileGitlab) reconcileRedisStatefulSet(cr *gitlabv1beta1.Gitlab) error {
	redis := getRedisStatefulSet(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: redis.Name}, redis) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, redis, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), redis)
}

func (r *ReconcileGitlab) reconcileGitalyStatefulSet(cr *gitlabv1beta1.Gitlab) error {
	gitaly := getGitalyStatefulSet(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: gitaly.Name}, gitaly) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, gitaly, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), gitaly)
}
