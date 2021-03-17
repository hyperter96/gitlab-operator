package gitlab

import (
	"fmt"
	"strings"

	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/helpers"
	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/helm"
)

const (
	// StrongPassword defines password length
	StrongPassword = 21

	// DatabaseName defines name of GitLab database in PostgreSQL
	DatabaseName = "gitlabhq_production"

	// DatabaseUser defines name of user in PostgreSQL
	DatabaseUser = "gitlab"

	// AppServiceAccount for GitLab app use
	AppServiceAccount = "gitlab-app"
)

// RedisConfig struct configures redis password
// and cluster configuration for large environments
type RedisConfig struct {
	// Defines the redis host to be used in the configuration
	Password string
	Cluster  bool
}

// ConfigurationOptions  holds
// options used to configure the different
// GitLab components
type ConfigurationOptions struct {
	// Namespace where the objects should live
	Namespace string

	// GitlabURL defines address reach deployed
	// Gitlab instance
	/* GitlabURL string */

	// RegistryURL defines web address to access
	// GitLab registry
	/* RegistryURL string */

	// PostgreSQL defines name of
	// database instance
	PostgreSQL string

	// EnableRegistry allows the user to disable the
	// GitLab container registry
	/* EnableRegistry bool */

	// Registry defines name of gitlab registry
	Registry string

	// ObjectStore defines object that describes values
	// for the S3 compatible storage service
	ObjectStore ObjectStoreOptions

	// Gitaly defines name of Gitaly server(s)
	Gitaly string

	// RedisMaster defines name of Redis instance
	RedisMaster string

	// Webservice defines name of the puma service which
	// listens on port 8181
	Webservice string

	// EmailFrom defines From address of outgoing email
	EmailFrom string

	// ReplyTo defines alternate email address to send admin emails
	ReplyTo string
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

	// Replicas for the development minio instance
	Replicas int32

	// VolumeSpec for the Minio instance
	/* gitlabv1beta1.VolumeSpec */
}

// SystemBuildOptions retrieves options from the Gitlab custom resource
// and uses them to build configuration options used to deploy
// the Gitlab instance
func SystemBuildOptions(cr *gitlabv1beta1.GitLab) ConfigurationOptions {

	values := helm.FromMap(cr.Spec.Chart.Values.Object)
	objectStoreEnabled, _ := helpers.GetBoolValue(values, "global.appConfig.object_store.enabled")
	objectStoreHost, _ := helpers.GetStringValue(values, "global.hosts.minio.name")
	// This implies that we only support global object-store config.
	objectStoreSecret, _ := helpers.GetStringValue(values, "global.appConfig.object_store.connection.secret")

	options := ConfigurationOptions{
		Namespace: cr.Namespace,
		/*
			GitlabURL:      DomainNameOnly(cr.Spec.URL),
			EnableRegistry: !cr.Spec.Registry.Disabled,
			RegistryURL:    DomainNameOnly(cr.Spec.Registry.URL),
		*/
		PostgreSQL:  getName(cr.Name, "postgresql"),
		RedisMaster: getName(cr.Name, "redis"),
		Gitaly:      getName(cr.Name, "gitaly"),
		Registry:    getName(cr.Name, "registry"),
		Webservice:  getName(cr.Name, "webservice"),
		ObjectStore: ObjectStoreOptions{
			URL:         objectStoreHost,
			Credentials: strings.Join([]string{cr.Name, "minio-secret"}, "-"),
			/*
				VolumeSpec: gitlabv1beta1.VolumeSpec{
					StorageClass: cr.Spec.ObjectStore.StorageClass,
				},
			*/
		},
	}

	// We can comment the following. SMTP options are not used.
	/*
		if IsEmailConfigured(cr) {
			options.EmailFrom, options.ReplyTo = setupSMTPOptions(cr)
		}
	*/

	if objectStoreEnabled {
		options.ObjectStore.URL = getName(cr.Name, "minio")
		options.ObjectStore.Capacity = "5Gi"
	}

	if objectStoreHost == "" {
		options.ObjectStore.Endpoint = ""
	}

	if objectStoreEnabled {
		minioSocket := []string{"http://", fmt.Sprintf("%s-minio", cr.Name), ":9000"}
		options.ObjectStore.Endpoint = strings.Join(minioSocket, "")
	} else {
		options.ObjectStore.Endpoint = fmt.Sprintf("https://%s", objectStoreHost)
	}

	if !objectStoreEnabled && options.ObjectStore.Credentials != "" {
		options.ObjectStore.Credentials = objectStoreSecret
	}

	return options
}

// RailsOptions defines parameters
// for rails secret
type RailsOptions struct {
	SecretKey     string
	DatabaseKey   string
	OTPKey        string
	RSAPrivateKey []string
	JWTSigningKey []string
}
