package gitlab

import (
	"context"
	"fmt"
	"net"
	"strings"

	gitlabv1beta1 "gitlab.com/ochienged/gitlab-operator/pkg/apis/gitlab/v1beta1"
	gitlabutils "gitlab.com/ochienged/gitlab-operator/pkg/controller/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GenerateComponentPasswords generates passwords for the
// different gitlab components
func (c *ComponentPasswords) GenerateComponentPasswords() {
	if c.redis == "" {
		c.redis = gitlabutils.Password(gitlabutils.PasswordOptions{
			EnableSpecialChars: false,
			Length:             StrongPassword,
		})
	}

	if c.postgres == "" {
		c.postgres = gitlabutils.Password(gitlabutils.PasswordOptions{
			EnableSpecialChars: false,
			Length:             StrongPassword,
		})
	}

	if c.runnerRegistrationToken == "" {
		c.runnerRegistrationToken = gitlabutils.Password(gitlabutils.PasswordOptions{
			EnableSpecialChars: false,
			Length:             StrongPassword,
		})
	}

	if c.gitlabRootPassword == "" {
		c.gitlabRootPassword = gitlabutils.Password(gitlabutils.PasswordOptions{
			EnableSpecialChars: true,
			Length:             StrongPassword,
		})
	}
}

// RunnerRegistrationToken returns password for linking runner to gitlab
func (c *ComponentPasswords) RunnerRegistrationToken() string {
	return c.runnerRegistrationToken
}

// GitlabRootPassword returns gitlab root password for web interface
func (c *ComponentPasswords) GitlabRootPassword() string {
	return c.gitlabRootPassword
}

// PostgresPassword returns postgres password
func (c *ComponentPasswords) PostgresPassword() string {
	return c.postgres
}

// RedisPassword returns redis password
func (c *ComponentPasswords) RedisPassword() string {
	return c.redis
}

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
	return cr.Spec.TLSCertificate != ""
}

// Function watches for database to startup.
// Watches endpoint for pod IP address
func isDatabaseReady(cr *gitlabv1beta1.Gitlab) bool {
	var addresses []corev1.EndpointAddress
	client, err := gitlabutils.NewKubernetesClient()
	if err != nil {
		log.Error(err, "Unable to acquire client")
	}

	endpoint, err := client.CoreV1().Endpoints(cr.Namespace).Get(cr.Name+"-database", metav1.GetOptions{})
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

// Get the Kubernetes service and pod network
// CIDRs for whitelisting
func getMonitoringWhitelist(cr *gitlabv1beta1.Gitlab) string {
	var addresses []corev1.EndpointAddress
	var networks []string
	endpoint, err := getOperatorMetricsEndpoints(cr)
	if err != nil {
		log.Error(err, "Error getting metrics endpoint")
	}

	for _, subset := range endpoint.Subsets {
		addresses = append(addresses, subset.Addresses...)
	}

	for _, address := range addresses {
		if address.IP != "" && !isAddressInList(networks, address.IP) {
			networks = append(networks, getNetworkAddress(address.IP))
		}
	}

	// TODO: Consider user provided whitelist
	return fmt.Sprintf("['%s']", strings.Join(networks, "', '"))
}

// Returns true if item is in slice
// false, otherwise
func isAddressInList(slice []string, key string) bool {
	if len(slice) == 0 {
		return false
	}

	for _, item := range slice {
		if item == key {
			return true
		}
	}

	return false
}

// SetStatus sets status of custom resource
func SetStatus(client client.Client, object runtime.Object) (err error) {
	err = client.Status().Update(context.TODO(), object)
	return
}

func gitalyShellConfigs(cr *gitlabv1beta1.Gitlab) GitalyConfig {
	return GitalyConfig{
		RedisService:   strings.Join([]string{cr.Name, "-redis"}, "-"),
		UnicornService: strings.Join([]string{cr.Name, "-unicorn"}, "-"),
	}
}

func getName(cr, component string) string {
	return strings.Join([]string{cr, component}, "-")
}
