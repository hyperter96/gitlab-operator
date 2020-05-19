package gitlab

import (
	"bytes"
	"fmt"
	"net"
	"reflect"
	"regexp"
	"strings"
	"text/template"

	gitlabv1beta1 "gitlab.com/ochienged/gitlab-operator/pkg/apis/gitlab/v1beta1"
	gitlabutils "gitlab.com/ochienged/gitlab-operator/pkg/controller/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DomainNameOnly separates domain from URL
func DomainNameOnly(url string) string {
	if strings.Contains(url, "://") {
		return strings.Split(url, "://")[1]
	}

	return url
}

func parseURL(url string, secured bool) string {
	var domain string
	protocol := "http"
	if strings.Contains(url, "://") {
		domain = strings.Split(url, "://")[1]
	}

	if domain != "" {
		return strings.Join([]string{"http://", domain}, "")
	}

	if secured {
		protocol = "https"
	}

	return strings.Join([]string{protocol, "://", url}, "")
}

func hasTLS(cr *gitlabv1beta1.Gitlab) bool {
	return cr.Spec.TLS != ""
}

// Function checks if endpoint is ready to accept connections.
// Watches endpoint for pod IP address
func isEndpointReady(service string, cr *gitlabv1beta1.Gitlab) bool {
	var addresses []corev1.EndpointAddress
	client, err := gitlabutils.NewKubernetesClient()
	if err != nil {
		log.Error(err, "Unable to acquire client")
	}

	endpoint, err := client.CoreV1().Endpoints(cr.Namespace).Get(service, metav1.GetOptions{})
	if err != nil {
		// Endpoint was not found so return false
		return false
	}

	for _, subset := range endpoint.Subsets {
		addresses = append(addresses, subset.Addresses...)
	}

	// If more than one IP address is returned,
	// The database is up and listening for connections
	return len(addresses) > 0
}

// The database endpoint returns the gitlab database endpoint
// Function is used by other functions e.g. isDatabaseReady
func getOperatorMetricsEndpoints(cr *gitlabv1beta1.Gitlab) (*corev1.Endpoints, error) {
	operatorMetricsSVC := "gitlab-operator-metrics"
	client, err := gitlabutils.NewKubernetesClient()
	if err != nil {
		log.Error(err, "Unable to acquire client")
	}

	return client.CoreV1().Endpoints(cr.Namespace).Get(operatorMetricsSVC, metav1.GetOptions{})
}

func getNetworkAddress(address string) string {
	if !strings.Contains(address, "/") {
		address = fmt.Sprintf("%s/8", address)
	}

	_, network, err := net.ParseCIDR(address)
	if err != nil {
		log.Error(err, "Unable to get network CIDR")
	}
	return network.String()
}

func getName(cr, component string) string {
	return strings.Join([]string{cr, component}, "-")
}

func getDatabaseConfiguration(postgres string) string {
	var config bytes.Buffer
	options := ConfigOptions{
		Postgres: postgres,
	}

	postgresTemplate := template.Must(template.ParseFiles("/templates/database.yml.erb"))
	postgresTemplate.Execute(&config, options)

	return config.String()
}

func getRedisConfiguration(redis string) string {
	var config bytes.Buffer
	options := ConfigOptions{
		RedisMaster: redis,
	}
	resqueTemplate := template.Must(template.ParseFiles("/templates/resque.yml.erb"))
	resqueTemplate.Execute(&config, options)

	return config.String()
}

// IsEmailConfigured returns true if SMTP is configured for the
// Gitlab resource. False, otherwise
func IsEmailConfigured(cr *gitlabv1beta1.Gitlab) bool {
	return !reflect.DeepEqual(cr.Spec.SMTP, gitlabv1beta1.SMTPConfiguration{})
}

// Return emailFrom and ReplyTo values
func setupSMTPOptions(cr *gitlabv1beta1.Gitlab) (emailFrom, replyTo string) {
	if cr.Spec.SMTP.EmailFrom != "" {
		emailFrom = cr.Spec.SMTP.EmailFrom
	} else if cr.Spec.SMTP.EmailFrom == "" && gitlabutils.IsEmailAddress(cr.Spec.SMTP.Username) {
		emailFrom = cr.Spec.SMTP.Username
	}

	if cr.Spec.SMTP.ReplyTo != "" {
		replyTo = cr.Spec.SMTP.ReplyTo
	} else if cr.Spec.SMTP.ReplyTo == "" && gitlabutils.IsEmailAddress(cr.Spec.SMTP.Username) {
		replyTo = cr.Spec.SMTP.Username
	}
	return
}

// This function checks the GitLab kind SMTP options and
// returns an smtp_settings.rb file in string format
func getSMTPSettings(cr *gitlabv1beta1.Gitlab) string {
	var settings bytes.Buffer

	if reflect.DeepEqual(cr.Spec.SMTP, gitlabv1beta1.SMTPConfiguration{}) {
		return ""
	}

	smtpTemplate := template.Must(template.ParseFiles("/templates/smtp_settings.rb"))
	smtpTemplate.Execute(&settings, cr.Spec.SMTP)

	// Remove whitespaces
	pattern := regexp.MustCompile(`(?m)^\s+[\n\r]+|[\r\n]+\s+\z`)

	return pattern.ReplaceAllString(settings.String(), "\n")
}

func getPostgresOverrides(postgres *gitlabv1beta1.DatabaseSpec) gitlabv1beta1.DatabaseSpec {
	if postgres != nil {
		return *postgres
	}

	var replicas int32 = 1
	return gitlabv1beta1.DatabaseSpec{
		Replicas: replicas,
	}
}

func getRedisOverrides(redis *gitlabv1beta1.RedisSpec) gitlabv1beta1.RedisSpec {
	if redis != nil {
		return *redis
	}

	var replicas int32 = 1
	return gitlabv1beta1.RedisSpec{
		Replicas: replicas,
	}
}

func getMinioOverrides(minio *gitlabv1beta1.MinioSpec) gitlabv1beta1.MinioSpec {
	if minio != nil {
		if minio.Volume.Capacity == "" {
			minio.Volume.Capacity = "5Gi"
		}
		return *minio
	}

	var replicas int32 = 4
	return gitlabv1beta1.MinioSpec{
		Replicas: replicas,
		Volume: gitlabv1beta1.VolumeSpec{
			Capacity: "5Gi",
		},
	}
}

func getGitlabURL(cr *gitlabv1beta1.Gitlab) string {
	if cr.Spec.URL != "" {
		return DomainNameOnly(cr.Spec.URL)
	}

	return "gitlab.example.com"
}

func getRegistryURL(cr *gitlabv1beta1.Gitlab) string {
	if cr.Spec.Registry.URL != "" {
		return DomainNameOnly(cr.Spec.Registry.URL)
	}

	return "registry.example.com"
}

func getMinioURL(cr *gitlabv1beta1.Gitlab) string {
	minio := getMinioOverrides(cr.Spec.Minio)
	if cr.Spec.Minio != nil {
		return DomainNameOnly(minio.URL)
	}

	return "minio.example.com"
}
