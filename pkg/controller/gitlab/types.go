package gitlab

import (
	"fmt"
	"strings"

	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/pkg/apis/gitlab/v1beta1"
	gitlabutils "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/pkg/controller/utils"
	"k8s.io/apimachinery/pkg/api/errors"
)

const (
	// StrongPassword defines password length
	StrongPassword = 21

	// DatabaseName defines name of GitLab database in PostgreSQL
	DatabaseName = "gitlabhq_production"

	// DatabaseUser defines name of user in PostgreSQL
	DatabaseUser = "gitlab"
)

// RedisConfig struct configures redis password
// and cluster configuration for large environments
type RedisConfig struct {
	// Defines the redis host to be used in the configuration
	Password string
	Cluster  bool
}

// ReadinessStatus shows status of Gitlab services
type ReadinessStatus struct {
	// Returns status of Gitlab rails app
	WorkhorseStatus string `json:"status,omitempty"`
	// RedisStatus reports status of redis
	RedisStatus []ServiceStatus `json:"redis_check,omitempty"`
	// DatabaseStatus reports status of postgres
	DatabaseStatus []ServiceStatus `json:"db_check,omitempty"`
}

// ServiceStatus shows status of a Gitlab
// dependent service .e.g. Postgres, Redis, Gitaly
type ServiceStatus struct {
	Status string `json:"status,omitempty"`
}

// ConfigurationOptions  holds
// options used to configure the different
// GitLab components
type ConfigurationOptions struct {
	// Namespace where the objects should live
	Namespace string

	// GitlabURL defines address reach deployed
	// Gitlab instance
	GitlabURL string

	// RegistryURL defines web address to access
	// GitLab registry
	RegistryURL string

	// PostgreSQL defines name of
	// database instance
	PostgreSQL string

	// EnableRegistry allows the user to disable the
	// GitLab container registry
	EnableRegistry bool

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

	// AccessKey used to authenticate against s3 storage
	AccessKey string

	// SecretKey used to authenticate against s3 storage
	SecretKey string

	// Replicas for the development minio instance
	Replicas int32

	// VolumeSpec for the Minio instance
	gitlabv1beta1.VolumeSpec
}

// SystemBuildOptions retrieves options from the Gitlab custom resource
// and uses them to build configuration options used to deploy
// the Gitlab instance
func SystemBuildOptions(cr *gitlabv1beta1.Gitlab) ConfigurationOptions {
	options := ConfigurationOptions{
		Namespace:      cr.Namespace,
		GitlabURL:      DomainNameOnly(cr.Spec.URL),
		EnableRegistry: !cr.Spec.Registry.Disabled,
		RegistryURL:    DomainNameOnly(cr.Spec.Registry.URL),
		PostgreSQL:     getName(cr.Name, "postgresql"),
		RedisMaster:    getName(cr.Name, "redis"),
		Gitaly:         getName(cr.Name, "gitaly"),
		Registry:       getName(cr.Name, "registry"),
		Webservice:     getName(cr.Name, "webservice"),
		ObjectStore: ObjectStoreOptions{
			URL:         DomainNameOnly(cr.Spec.ObjectStore.URL),
			Credentials: strings.Join([]string{cr.Name, "minio-secret"}, "-"),
			VolumeSpec: gitlabv1beta1.VolumeSpec{
				StorageClass: cr.Spec.ObjectStore.StorageClass,
			},
		},
	}

	if IsEmailConfigured(cr) {
		options.EmailFrom, options.ReplyTo = setupSMTPOptions(cr)
	}

	if cr.Spec.ObjectStore.Development {
		options.ObjectStore.URL = getName(cr.Name, "minio")
		options.ObjectStore.Capacity = "5Gi"
	}

	setObjectStoreEndpoint(cr, &options)

	if cr.Spec.ObjectStore.Credentials != "" {
		getObjectStoreKeys(cr, &options)
	}

	return options
}

// set up the enpoint and othe options for the
// S3 object store service
func setObjectStoreEndpoint(cr *gitlabv1beta1.Gitlab, options *ConfigurationOptions) {
	var port string
	protocol := "https"

	if cr.Spec.ObjectStore.Development {
		protocol = "http"
		port = ":9000"
		options.ObjectStore.Endpoint = strings.Join([]string{fmt.Sprintf("%s://", protocol), options.ObjectStore.URL, port}, "")
	}

	if cr.Spec.ObjectStore.URL == "" {
		options.ObjectStore.Endpoint = ""
	}

	if strings.Contains(cr.Spec.ObjectStore.URL, "://") {
		options.ObjectStore.Endpoint = cr.Spec.ObjectStore.URL
	}

	// Sets the name of secret with 'accesskey' and 'secretkey'
	if credentials := cr.Spec.ObjectStore.Credentials; credentials != "" {
		options.ObjectStore.Credentials = credentials
	}

	options.ObjectStore.Endpoint = strings.Join([]string{fmt.Sprintf("%s://", protocol), cr.Spec.ObjectStore.URL}, "")
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

func getObjectStoreKeys(cr *gitlabv1beta1.Gitlab, options *ConfigurationOptions) {
	keys, err := gitlabutils.SecretData(cr.Spec.ObjectStore.Credentials, cr.Namespace)
	if err != nil && errors.IsNotFound(err) {
		log.Error(err, "Invalid object store credentials")
	}

	options.ObjectStore.AccessKey = keys["accesskey"]
	options.ObjectStore.SecretKey = keys["secretkey"]
}
