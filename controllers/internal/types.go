package internal

import (
	"fmt"
	"strings"

	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/gitlab"
)

const (
	// GitlabType represents resource of type Gitlab
	GitlabType = "gitlab"
)

var (
	// ConfigMapDefaultMode for configmap projected volume
	ConfigMapDefaultMode int32 = 420

	// ExecutableDefaultMode for configmap projected volume
	ExecutableDefaultMode int32 = 493

	// ProjectedVolumeDefaultMode for projected volume
	ProjectedVolumeDefaultMode int32 = 256

	// SecretDefaultMode for secret projected volume
	SecretDefaultMode int32 = 288
)

// PasswordOptions provides parameters to be
// used when generating passwords
type PasswordOptions struct {
	// Length defines desired password length
	Length int
	// EnableSpecialCharacters adds special characters
	// to generated passwords
	EnableSpecialChars bool
}

// ConfigurationOptions holds options used to configure the different
// GitLab components
type ConfigurationOptions struct {
	// ObjectStore defines object that describes values
	// for the S3 compatible storage service
	ObjectStore ObjectStoreOptions
}

// ObjectStoreOptions defines properties for
// the S3 storage used by GitLab
type ObjectStoreOptions struct {
	// URL defines address for development
	// S3 storage service
	URL string

	// Endpoint defines the URL the API endpoint
	// including the protocol
	Endpoint string

	// Credentials is the name of the secret
	// that contains the 'accesskey' and 'secretkey'
	Credentials string

	// Capacity of the volume to be used by the development
	// minio instance
	Capacity string
}

// SystemBuildOptions retrieves options from the Gitlab custom resource
// and uses them to build configuration options used to deploy
// the Gitlab instance
func SystemBuildOptions(adapter gitlab.CustomResourceAdapter) ConfigurationOptions {
	objectStoreEnabled, _ := gitlab.GetBoolValue(adapter.Values(), "global.appConfig.object_store.enabled")
	objectStoreHost, _ := gitlab.GetStringValue(adapter.Values(), "global.hosts.minio.name")
	// This implies that we only support global object-store config.
	objectStoreSecret, _ := gitlab.GetStringValue(adapter.Values(), "global.appConfig.object_store.connection.secret")

	options := ConfigurationOptions{
		ObjectStore: ObjectStoreOptions{
			URL:         objectStoreHost,
			Credentials: strings.Join([]string{adapter.ReleaseName(), "minio-secret"}, "-"),
		},
	}

	objectStoreCapacity, _ := gitlab.GetStringValue(adapter.Values(), "minio.persistence.size")
	if objectStoreCapacity == "" {
		objectStoreCapacity = "5Gi"
	}

	if objectStoreEnabled {
		options.ObjectStore.URL = getName(adapter.ReleaseName(), "minio")
		options.ObjectStore.Capacity = objectStoreCapacity
	}

	if objectStoreHost == "" {
		options.ObjectStore.Endpoint = ""
	}

	if objectStoreEnabled {
		minioSocket := []string{"http://", fmt.Sprintf("%s-minio", adapter.ReleaseName()), ":9000"}
		options.ObjectStore.Endpoint = strings.Join(minioSocket, "")
	} else {
		options.ObjectStore.Endpoint = fmt.Sprintf("https://%s", objectStoreHost)
	}

	if !objectStoreEnabled && options.ObjectStore.Credentials != "" {
		options.ObjectStore.Credentials = objectStoreSecret
	}

	return options
}
