package internal

import (
	"encoding/json"
	"fmt"
	"os"

	yaml "gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/gitlab"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/settings"
)

var (
	localUser int64 = 1000
)

// MinioSecret returns secret containing Minio accesskey and secretkey.
func MinioSecret(adapter gitlab.CustomResourceAdapter) *corev1.Secret {
	labels := Label(adapter.ReleaseName(), "minio", GitlabType)
	options := SystemBuildOptions(adapter)

	secretKey := Password(PasswordOptions{
		EnableSpecialChars: false,
		Length:             48,
	})

	minio := GenericSecret(options.ObjectStore.Credentials, adapter.Namespace(), labels)

	minio.Data = map[string][]byte{
		"accesskey": []byte("gitlab"),
		"secretkey": []byte(secretKey),
	}

	return minio
}

// MinioStatefulSet return Minio statefulset.
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

// MinioScriptConfigMap returns scripts used to configure Minio.
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

// MinioIngress returns the Ingress that exposes MinIO.
func MinioIngress(adapter gitlab.CustomResourceAdapter) *networkingv1.Ingress {
	ingressClass, err := gitlab.GetStringValue(adapter.Values(), "global.ingress.class")
	if err != nil || ingressClass == "" {
		ingressClass = fmt.Sprintf("%s-nginx", adapter.ReleaseName())
	}

	labels := Label(adapter.ReleaseName(), "minio", GitlabType)
	annotations := map[string]string{
		"kubernetes.io/ingress.class":                         ingressClass,
		"kubernetes.io/ingress.provider":                      "nginx",
		"nginx.ingress.kubernetes.io/proxy-body-size":         "0",
		"nginx.ingress.kubernetes.io/proxy-buffering":         "off",
		"nginx.ingress.kubernetes.io/proxy-read-timeout":      "900",
		"nginx.ingress.kubernetes.io/proxy-request-buffering": "off",
	}

	configureCertmanager := CertManagerEnabled(adapter)

	if configureCertmanager {
		annotations["cert-manager.io/issuer"] = fmt.Sprintf("%s-issuer", adapter.ReleaseName())
		annotations["acme.cert-manager.io/http01-edit-in-place"] = "true"
	}

	minioIngressTLSSecretName, _ := gitlab.GetStringValue(adapter.Values(), "minio.ingress.tls.secretName")
	globalIngressTLSSecretName, _ := gitlab.GetStringValue(adapter.Values(), "global.ingress.tls.secretName")
	wildcardIngressTLSSecretName := fmt.Sprintf("%s-wildcard-tls", adapter.ReleaseName())
	certmanagerIngressTLSSecretName := fmt.Sprintf("%s-minio-tls", adapter.ReleaseName())

	// To determine the secret name, copy the 'minio.tlsSecret' template logic until MinIO generated resources are removed.
	// https://gitlab.com/gitlab-org/charts/gitlab/blob/8bc9d9e5f883e6d750cc957d4c1c67e65a9222df/charts/minio/templates/_helpers.tpl#L41-54
	var tlsSecretName string
	if minioIngressTLSSecretName != "" {
		tlsSecretName = minioIngressTLSSecretName
	} else if globalIngressTLSSecretName != "" {
		tlsSecretName = globalIngressTLSSecretName
	} else if configureCertmanager {
		tlsSecretName = certmanagerIngressTLSSecretName
	} else {
		tlsSecretName = wildcardIngressTLSSecretName
	}

	url := getMinioURL(adapter)
	pathType := networkingv1.PathTypePrefix

	return &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        labels["app.kubernetes.io/instance"],
			Namespace:   adapter.Namespace(),
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: url,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: MinioService(adapter).Name,
											Port: networkingv1.ServiceBackendPort{
												Number: 9000,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			TLS: []networkingv1.IngressTLS{
				{
					Hosts:      []string{url},
					SecretName: tlsSecretName,
				},
			},
		},
	}
}

// MinioService returns service that exposes Minio.
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
					Name:       "minio",
					Port:       9000,
					TargetPort: intstr.FromInt(9000),
					Protocol:   corev1.ProtocolTCP,
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

	data := minioSecret.Data

	connectionInfo := map[string]string{
		"provider":              "AWS",
		"region":                settings.Region,
		"host":                  options.ObjectStore.URL,
		"endpoint":              options.ObjectStore.Endpoint,
		"aws_access_key_id":     string(data["accesskey"]),
		"aws_secret_access_key": string(data["secretkey"]),
		"path_style":            "true",
	}

	connectionBytes, err := json.Marshal(connectionInfo)
	if err != nil {
		return &corev1.Secret{}, fmt.Errorf("unable to encode connection string for storage-config")
	}

	secret.Data = map[string][]byte{
		"connection": connectionBytes,
	}

	return secret, nil
}

// RegistryConnectionSecret returns secret containing MinIO connection config for `registry.storage.secret`.
func RegistryConnectionSecret(adapter gitlab.CustomResourceAdapter, minioSecret corev1.Secret) (*corev1.Secret, error) {
	labels := Label(adapter.ReleaseName(), "minio", GitlabType)
	options := SystemBuildOptions(adapter)
	secret := GenericSecret(settings.RegistryConnectionSecretName, adapter.Namespace(), labels)

	data := minioSecret.Data

	minioEndpoint := options.ObjectStore.Endpoint
	minioRedirect, _ := gitlab.GetBoolValue(adapter.Values(), "registry.minio.redirect", false)

	if minioRedirect {
		// We can always assume it is HTTPS.
		minioEndpoint = fmt.Sprintf("https://%s", getMinioURL(adapter))
	}

	connectionInfo := map[string]map[string]string{
		"s3": {
			"bucket":         settings.RegistryBucket,
			"accesskey":      string(data["accesskey"]),
			"secretkey":      string(data["secretkey"]),
			"region":         settings.Region,
			"regionendpoint": minioEndpoint,
			"v4auth":         "true",
			"path_style":     "true",
		},
	}

	connectionBytes, err := yaml.Marshal(connectionInfo)
	if err != nil {
		return &corev1.Secret{}, fmt.Errorf("unable to encode connection string for registry")
	}

	secret.Data = map[string][]byte{
		"config": connectionBytes,
	}

	return secret, nil
}

// ToolboxConnectionSecret returns secret containing MinIO connection config for `global.toolbox.backups.objectStorage.config.secret`.
func ToolboxConnectionSecret(adapter gitlab.CustomResourceAdapter, minioSecret corev1.Secret) *corev1.Secret {
	labels := Label(adapter.ReleaseName(), "minio", GitlabType)
	secret := GenericSecret(settings.ToolboxConnectionSecretName, adapter.Namespace(), labels)
	url := getMinioURL(adapter)
	data := minioSecret.Data

	template := `
[default]
access_key = %s
secret_key = %s
bucket_location = %s
host_base = %s
host_bucket = %s/%%(bucket)
default_mime_type = binary/octet-stream
enable_multipart = True
multipart_max_chunks = 10000
multipart_chunk_size_mb = 128
recursive = True
recv_chunk = 65536
send_chunk = 65536
server_side_encryption = False
signature_v2 = True
socket_timeout = 300
use_mime_magic = False
verbosity = WARNING
website_endpoint = https://%s
`

	result := fmt.Sprintf(template, data["accesskey"], data["secretkey"], settings.Region, url, url, url)

	secret.Data = map[string][]byte{
		"config": []byte(result),
	}

	return secret
}

func getMinioURL(adapter gitlab.CustomResourceAdapter) string {
	name, _ := gitlab.GetStringValue(adapter.Values(), "global.hosts.minio.name")
	if name != "" {
		return name
	}

	hostSuffix, _ := gitlab.GetStringValue(adapter.Values(), "global.hosts.hostSuffix")
	domain, _ := gitlab.GetStringValue(adapter.Values(), "global.hosts.domain", "example.com")

	if hostSuffix != "" {
		return fmt.Sprintf("minio-%s.%s", hostSuffix, domain)
	}

	return fmt.Sprintf("minio.%s", domain)
}
