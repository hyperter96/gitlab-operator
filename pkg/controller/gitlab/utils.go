package gitlab

import (
	"math/rand"
	"strings"
	"time"

	gitlabv1beta1 "github.com/OchiengEd/gitlab-operator/pkg/apis/gitlab/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// GeneratePassword creates an alphanumeric
// string to be used as a password
func GeneratePassword(options PasswordOptions) string {
	password := make([]byte, options.Length)
	charset := []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz01234567890")
	special := []byte("+,#?$()&%^*=/!+<>[]{}@-_")

	// add special characters if not database password
	if !options.Database {
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
func GetSecretValue(namespace, secret, key string) []byte {
	client, err := NewKubernetesClient()
	if err != nil {
		log.Error(err, "Unable to get kubernetes client")
	}

	store, err := client.CoreV1().Secrets(namespace).Get(secret, metav1.GetOptions{})
	if err != nil {
		log.Error(err, "Secret not found")
	}
	return store.Data[key]
}

// GeneratePasswords generates passwords for the different gitlab
// components
func (c *ComponentPasswords) GeneratePasswords() {
	if c.redis == "" {
		c.redis = GeneratePassword(PasswordOptions{Database: false, Length: StrongPassword})
	}

	if c.postgres == "" {
		c.postgres = GeneratePassword(PasswordOptions{Database: true, Length: StrongPassword})
	}

	if c.runnerRegistrationToken == "" {
		c.runnerRegistrationToken = GeneratePassword(PasswordOptions{Database: false, Length: StrongPassword})
	}

	if c.gitlabRootPassword == "" {
		c.gitlabRootPassword = GeneratePassword(PasswordOptions{Database: false, Length: StrongPassword})
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

func getLabels(cr *gitlabv1beta1.Gitlab, component string) map[string]string {
	var edition string = "community"

	if cr.Spec.Enterprise {
		edition = "enterprise"
	}

	return map[string]string{
		"app.kubernetes.io/name":       cr.Name + "-" + component,
		"app.kubernetes.io/component":  component,
		"app.kubernetes.io/edition":    edition,
		"app.kubernetes.io/part-of":    "gitlab",
		"app.kubernetes.io/managed-by": "gitlab-operator",
	}

}

func getVolumeRequest(size string) corev1.ResourceList {
	return corev1.ResourceList{
		"storage": resource.MustParse(size),
	}
}

// GetDomainNameOnly separates domain from URL
func GetDomainNameOnly(url string) string {
	var domain []string
	if strings.Contains(url, "://") {
		domain = strings.Split(url, "://")
	}

	return domain[1]
}
