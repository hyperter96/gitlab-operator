package internal

import (
	"encoding/json"
	"fmt"
	"os"

	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/gitlab"
	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/settings"
	yaml "gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// MinioSecret returns secret containing Minio accesskey and secretkey
func MinioSecret(adapter gitlab.CustomResourceAdapter) *corev1.Secret {
	labels := Label(adapter.ReleaseName(), "minio", GitlabType)
	options := SystemBuildOptions(adapter)

	secretKey := Password(PasswordOptions{
		EnableSpecialChars: false,
		Length:             48,
	})

	minio := GenericSecret(options.ObjectStore.Credentials, adapter.Namespace(), labels)
	minio.StringData = map[string]string{
		"accesskey": "gitlab",
		"secretkey": secretKey,
	}

	return minio
}

// MinioStatefulSet return Minio statefulset
func MinioStatefulSet(adapter gitlab.CustomResourceAdapter) *appsv1.StatefulSet {
	labels := Label(adapter.ReleaseName(), "minio", GitlabType)
	options := SystemBuildOptions(adapter)

	var replicas int32 = 1

	minio := GenericStatefulSet(Component{
		Namespace: adapter.Namespace(),
		Labels:    labels,
		Replicas:  replicas,
		InitContainers: []corev1.Container{
			{
				Name:            "configure",
				Image:           "busybox:latest",
				ImagePullPolicy: corev1.PullIfNotPresent,
				Command:         []string{"sh", "/config/configure"},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						"cpu": ResourceQuantity("50m"),
					},
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "minio-configuration",
						MountPath: "/config",
					},
					{
						Name:      "minio-server-config",
						MountPath: "/minio",
					},
				},
			},
		},
		Containers: []corev1.Container{
			{
				Name:            "minio",
				Image:           "minio/minio:RELEASE.2017-12-28T01-21-00Z",
				ImagePullPolicy: corev1.PullIfNotPresent,
				Args:            []string{"-C", "/tmp/.minio", "--quiet", "server", "/export"},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						"cpu":    ResourceQuantity("100m"),
						"memory": ResourceQuantity("128Mi"),
					},
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "export",
						MountPath: "/export",
					},
					{
						Name:      "minio-server-config",
						MountPath: "/tmp/.minio",
					},
					{
						Name:      "podinfo",
						MountPath: "/podinfo",
						ReadOnly:  false,
					},
				},
				Ports: []corev1.ContainerPort{
					{
						Name:          "service",
						Protocol:      corev1.ProtocolTCP,
						ContainerPort: 9000,
					},
				},
				LivenessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						TCPSocket: &corev1.TCPSocketAction{
							Port: intstr.IntOrString{
								IntVal: 9000,
							},
						},
					},
					TimeoutSeconds: 1,
				},
			},
		},
		Volumes: []corev1.Volume{
			{
				Name: "podinfo",
				VolumeSource: corev1.VolumeSource{
					DownwardAPI: &corev1.DownwardAPIVolumeSource{
						Items: []corev1.DownwardAPIVolumeFile{
							{
								Path: "labels",
								FieldRef: &corev1.ObjectFieldSelector{
									FieldPath: "metadata.labels",
								},
							},
						},
					},
				},
			},
			{
				Name: "minio-server-config",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{
						Medium: corev1.StorageMediumMemory,
					},
				},
			},
			{
				Name: "minio-configuration",
				VolumeSource: corev1.VolumeSource{
					Projected: &corev1.ProjectedVolumeSource{
						Sources: []corev1.VolumeProjection{
							{
								ConfigMap: &corev1.ConfigMapProjection{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: adapter.ReleaseName() + "-minio-script",
									},
								},
							},
							{
								Secret: &corev1.SecretProjection{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: options.ObjectStore.Credentials,
									},
								},
							},
						},
					},
				},
			},
		},
		VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "export",
					Namespace: adapter.Namespace(),
					Labels:    labels,
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteOnce,
					},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							"storage": ResourceQuantity(options.ObjectStore.Capacity),
						},
					},
				},
			},
		},
	})

	minio.Spec.Template.Spec.SecurityContext = &corev1.PodSecurityContext{
		RunAsUser: &localUser,
		FSGroup:   &localUser,
	}

	minio.Spec.Template.Spec.ServiceAccountName = settings.AppServiceAccount

	return minio
}

// MinioScriptConfigMap returns scripts used to configure Minio
func MinioScriptConfigMap(adapter gitlab.CustomResourceAdapter) *corev1.ConfigMap {
	labels := Label(adapter.ReleaseName(), "minio", GitlabType)

	initScript := ReadConfig(os.Getenv("GITLAB_OPERATOR_ASSETS") + "/templates/minio/initialize-buckets.sh")
	configureScript := ReadConfig(os.Getenv("GITLAB_OPERATOR_ASSETS") + "/templates/minio/configure.sh")
	configJSON := ReadConfig(os.Getenv("GITLAB_OPERATOR_ASSETS") + "/templates/minio/config.json")

	init := GenericConfigMap(adapter.ReleaseName()+"-minio-script", adapter.Namespace(), labels)
	init.Data = map[string]string{
		"initialize":  initScript,
		"configure":   configureScript,
		"config.json": configJSON,
	}

	return init
}

// MinioService returns service that exposes Minio
func MinioService(adapter gitlab.CustomResourceAdapter) *corev1.Service {
	labels := Label(adapter.ReleaseName(), "minio", GitlabType)

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labels["app.kubernetes.io/instance"],
			Namespace: adapter.Namespace(),
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name:     "minio",
					Port:     9000,
					Protocol: corev1.ProtocolTCP,
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
}

// AppConfigConnectionSecret returns secret containing MinIO connection config for `global.appConfig.object_store.connection.secret`.
func AppConfigConnectionSecret(adapter gitlab.CustomResourceAdapter, minioSecret corev1.Secret) (*corev1.Secret, error) {
	labels := Label(adapter.ReleaseName(), "minio", GitlabType)
	options := SystemBuildOptions(adapter)
	secret := GenericSecret(settings.AppConfigConnectionSecretName, adapter.Namespace(), labels)

	data := minioSecret.StringData

	connectionInfo := map[string]string{
		"provider":              "AWS",
		"region":                settings.Region,
		"host":                  options.ObjectStore.URL,
		"endpoint":              options.ObjectStore.Endpoint,
		"aws_access_key_id":     data["accesskey"],
		"aws_secret_access_key": data["secretkey"],
	}

	connectionBytes, err := json.Marshal(connectionInfo)
	if err != nil {
		return &corev1.Secret{}, fmt.Errorf("unable to encode connection string for storage-config")
	}

	secret.StringData = map[string]string{
		"connection": string(connectionBytes),
	}

	return secret, nil
}

// RegistryConnectionSecret returns secret containing MinIO connection config for `registry.storage.secret`.
func RegistryConnectionSecret(adapter gitlab.CustomResourceAdapter, minioSecret corev1.Secret) (*corev1.Secret, error) {
	labels := Label(adapter.ReleaseName(), "minio", GitlabType)
	options := SystemBuildOptions(adapter)
	secret := GenericSecret(settings.RegistryConnectionSecretName, adapter.Namespace(), labels)

	data := minioSecret.StringData

	connectionInfo := map[string]map[string]string{
		"s3": {
			"bucket":         settings.RegistryBucket,
			"accesskey":      data["accesskey"],
			"secretkey":      data["secretkey"],
			"region":         settings.Region,
			"regionendpoint": options.ObjectStore.Endpoint,
			"v4auth":         "true",
		},
	}

	connectionBytes, err := yaml.Marshal(connectionInfo)
	if err != nil {
		return &corev1.Secret{}, fmt.Errorf("unable to encode connection string for registry")
	}

	secret.StringData = map[string]string{
		"config": string(connectionBytes),
	}

	return secret, nil
}

// TaskRunnerConnectionSecret returns secret containing MinIO connection config for `global.task-runner.backups.objectStorage.config.secret`.
func TaskRunnerConnectionSecret(adapter gitlab.CustomResourceAdapter, minioSecret corev1.Secret) *corev1.Secret {
	labels := Label(adapter.ReleaseName(), "minio", GitlabType)
	secret := GenericSecret(settings.TaskRunnerConnectionSecretName, adapter.Namespace(), labels)

	data := minioSecret.StringData

	template := `
	[default]
	access_key = %s
	secret_key = %s
	bucket_location = %s
	multipart_chunk_size_mb = 128 # default is 15 (MB)
	`

	result := fmt.Sprintf(template, data["accesskey"], data["secretkey"], settings.Region)

	secret.StringData = map[string]string{
		"config": result,
	}

	return secret
}