package gitlab

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"

	gitlabv1beta1 "github.com/OchiengEd/gitlab-operator/pkg/apis/gitlab/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GeneratePassword creates an alphanumeric
// string to be used as a password
func GeneratePassword(options PasswordOptions) string {
	password := make([]byte, options.Length)
	charset := []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz01234567890")
	special := []byte("+,#?$()&%^*=/!+<>[]{}@-_")

	// add special characters if not database password
	if options.EnableSpecialChars {
		charset = append(charset, special...)
	}

	seed := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := range password {
		password[i] = charset[seed.Intn(len(charset))]
	}

	return string(password)
}

// IsOpenshift check if API has the API route.openshift.io/v1,
// then it is considered an openshift environment
func IsOpenshift() bool {
	client, err := NewKubernetesClient()
	if err != nil {
		log.Error(err, "Unable to get kubernetes client")
	}

	routeGV := schema.GroupVersion{
		Group:   "route.openshift.io",
		Version: "v1",
	}

	if err := discovery.ServerSupportsVersion(client, routeGV); err != nil {
		return false
	}

	return true
}

// NewKubernetesClient returns a client that can be used to interact
// with the kubernetes api
func NewKubernetesClient() (clientset *kubernetes.Clientset, err error) {
	conf, err := rest.InClusterConfig()
	if err != nil {
		log.Error(err, "Unable getting cluster config")
	}

	clientset, err = kubernetes.NewForConfig(conf)
	return
}

// GetSecretValue returns the value for a key from an existing secret
func GetSecretValue(client client.Client, namespace, secret, key string) string {
	target := &corev1.Secret{}

	err := client.Get(context.TODO(), types.NamespacedName{Name: secret, Namespace: namespace}, target)
	if err != nil {
		log.Error(err, "Secret not found")
	}
	return string(target.Data[key])
}

// GenerateComponentPasswords generates passwords for the
// different gitlab components
func (c *ComponentPasswords) GenerateComponentPasswords() {
	if c.redis == "" {
		c.redis = GeneratePassword(PasswordOptions{
			EnableSpecialChars: false,
			Length:             StrongPassword,
		})
	}

	if c.postgres == "" {
		c.postgres = GeneratePassword(PasswordOptions{
			EnableSpecialChars: false,
			Length:             StrongPassword,
		})
	}

	if c.runnerRegistrationToken == "" {
		c.runnerRegistrationToken = GeneratePassword(PasswordOptions{
			EnableSpecialChars: false,
			Length:             StrongPassword,
		})
	}

	if c.gitlabRootPassword == "" {
		c.gitlabRootPassword = GeneratePassword(PasswordOptions{
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

func getLabels(cr *gitlabv1beta1.Gitlab, component string) (labels map[string]string) {
	var edition string = "community"

	if cr.Spec.Enterprise {
		edition = "enterprise"
	}

	labels = map[string]string{
		"app.kubernetes.io/name":       cr.Name,
		"app.kubernetes.io/instance":   strings.Join([]string{cr.Name, component}, "-"),
		"app.kubernetes.io/component":  component,
		"app.kubernetes.io/edition":    edition,
		"app.kubernetes.io/part-of":    "gitlab",
		"app.kubernetes.io/managed-by": "gitlab-operator",
	}

	return
}

func getVolumeRequest(size string) corev1.ResourceList {
	return corev1.ResourceList{
		"storage": resource.MustParse(size),
	}
}

// IsObjectFound checks if kubernetes resource is in the cluster
func IsObjectFound(client client.Client, key types.NamespacedName, object runtime.Object) bool {
	if err := client.Get(context.TODO(), key, object); err != nil {
		return false
	}

	return true
}

// GetDomainNameOnly separates domain from URL
func GetDomainNameOnly(url string) string {
	var domain []string
	if strings.Contains(url, "://") {
		domain = strings.Split(url, "://")
	}

	return domain[1]
}

// Function watches for database to startup.
// Watches endpoint for pod IP address
func isDatabaseReady(cr *gitlabv1beta1.Gitlab) bool {
	var addresses []corev1.EndpointAddress
	client, err := NewKubernetesClient()
	if err != nil {
		log.Error(err, "Unable to acquire client")
	}

	endpoint, err := client.CoreV1().Endpoints(cr.Namespace).Get(cr.Name+"-database", metav1.GetOptions{})
	if err != nil {
		fmt.Printf("Endpoint error; %v", err)
	}
	for _, subset := range endpoint.Subsets {
		addresses = append(addresses, subset.Addresses...)
	}

	// If more than one IP address is returned,
	// The database is up and listening for connections
	if len(addresses) > 0 {
		return true
	}

	return false
}

// The database endpoint returns the gitlab database endpoint
// Function is used by other functions e.g. isDatabaseReady
func getOperatorMetricsEndpoints(cr *gitlabv1beta1.Gitlab) (*corev1.Endpoints, error) {
	operatorMetricsSVC := "gitlab-operator-metrics"
	client, err := NewKubernetesClient()
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
		if address.IP != "" && !isPresent(networks, address.IP) {
			networks = append(networks, getNetworkAddress(address.IP))
		}
	}

	// TODO: Consider user provided whitelist
	return fmt.Sprintf("['%s']", strings.Join(networks, "', '"))
}

// Returns true if item is in slice
// false, otherwise
func isPresent(slice []string, key string) bool {
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

// IsPrometheusSupported checks for Prometheus API endpoint
func IsPrometheusSupported() bool {
	client, err := NewKubernetesClient()
	if err != nil {
		log.Error(err, "Unable to acquire k8s client")
	}

	servicemonGV := schema.GroupVersion{
		Group:   "monitoring.coreos.com",
		Version: "v1",
	}

	if err := discovery.ServerSupportsVersion(client, servicemonGV); err != nil {
		return false
	}

	return true
}
