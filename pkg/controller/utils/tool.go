package utils

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"encoding/base64"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Label function returns uniform labels for resources
func Label(resource, component, resourceType string) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":       resource,
		"app.kubernetes.io/instance":   strings.Join([]string{resource, component}, "-"),
		"app.kubernetes.io/component":  component,
		"app.kubernetes.io/part-of":    resourceType,
		"app.kubernetes.io/managed-by": "gitlab-operator",
	}
}

// ResourceLabels type encapsulates map[string]string
type ResourceLabels map[string]string

// GetName returns the name of resource based on instance label
func (l ResourceLabels) GetName() string {
	return l["app.kubernetes.io/instance"]
}

// ResourceQuantity returns quantity requested for resource
func ResourceQuantity(request string) resource.Quantity {
	return resource.MustParse(request)
}

// IsPrometheusSupported checks for Prometheus API endpoint
func IsPrometheusSupported() bool {
	client, err := KubernetesConfig().NewKubernetesClient()
	if err != nil {
		fmt.Printf("Unable to acquire k8s client: %v", err)
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

// IsOpenshift check if API has the API route.openshift.io/v1,
// then it is considered an openshift environment
func IsOpenshift() bool {
	client, err := KubernetesConfig().NewKubernetesClient()
	if err != nil {
		fmt.Printf("Unable to get kubernetes client: %v", err)
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

// IsObjectFound checks if kubernetes namespaced resource exists in the cluster
func IsObjectFound(client client.Client, key types.NamespacedName, object runtime.Object) bool {
	if err := client.Get(context.TODO(), key, object); err != nil {
		return false
	}

	return true
}

// KubeConfig returns kubernetes client configuration
type KubeConfig struct {
	Config *rest.Config
	Error  error
}

// KubernetesConfig returns kubernetes client config
func KubernetesConfig() KubeConfig {
	config, err := rest.InClusterConfig()
	if err != nil {
		return KubeConfig{
			Config: nil,
			Error:  err,
		}
	}

	return KubeConfig{
		Config: config,
		Error:  nil,
	}
}

// NewKubernetesClient returns a client that can be
// used to interact with the kubernetes api
func (k KubeConfig) NewKubernetesClient() (clientset *kubernetes.Clientset, err error) {
	conf := k.Config
	err = k.Error
	if err != nil {
		fmt.Printf("Error getting cluster config: %v", err)
	}

	clientset, err = kubernetes.NewForConfig(conf)
	return
}

// GetSecretValue returns the value for a key from an existing secret
func GetSecretValue(client client.Client, namespace, secret, key string) (string, error) {
	target := &corev1.Secret{}

	err := client.Get(context.TODO(), types.NamespacedName{Name: secret, Namespace: namespace}, target)
	if err != nil {
		return "", err
	}
	return string(target.Data[key]), err
}

// ReadConfig returns contents of file as string
func ReadConfig(filename string) string {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		// log.Error(err, "Error reading %s", filename)
		fmt.Printf("Error reading %s: %v", filename, err)
	}
	return string(content)
}

// RemoveEmptyLines removes empty lines from block string
func RemoveEmptyLines(block string) string {
	re, err := regexp.Compile("\n\n")
	if err != nil {
		fmt.Printf("Error building pattern: %v", err)
	}

	return re.ReplaceAllString(block, "\n")
}

// EncodeString accepts strings and returns a base64 encoded string
func EncodeString(message string) string {
	return base64.StdEncoding.EncodeToString([]byte(message))
}

// Password creates an password string based
// on the password options provided
func Password(options PasswordOptions) string {
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

// IsEmailAddress evaluates a string and returns true
// if a valid email address.
// False is otherwise returned
func IsEmailAddress(email string) bool {
	re := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.-]+\.[a-z.]{2,4}$`)
	return re.MatchString(email)
}

// IsPodRunning function tells user if pod is running
func IsPodRunning(pod *corev1.Pod) bool {
	return pod.Status.Phase == "Running"
}

// SecretData gets a secret by name and returns its data
func SecretData(name, namespace string) (map[string]string, error) {
	client, err := KubernetesConfig().NewKubernetesClient()
	if err != nil {
		return map[string]string{}, err
	}

	secret, err := client.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return map[string]string{}, err
	}

	return secret.StringData, nil
}

// IsMinioAvailable checks if Minio API provided
// by minio operator is present
func IsMinioAvailable() bool {
	client, err := KubernetesConfig().NewKubernetesClient()
	if err != nil {
		fmt.Printf("Unable to get kubernetes client: %v", err)
	}

	routeGV := schema.GroupVersion{
		Group:   "miniooperator.min.io",
		Version: "v1beta1",
	}

	if err := discovery.ServerSupportsVersion(client, routeGV); err != nil {
		return false
	}

	return true
}

// GetDeploymentPods returns the pods that belong to a given deployment
func GetDeploymentPods(kclient client.Client, name, namespace string) (result []corev1.Pod, err error) {
	deployment := &appsv1.Deployment{}
	err = kclient.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, deployment)
	if err != nil && errors.IsNotFound(err) {
		return result, err
	}

	deployLabels := deployment.Spec.Template.ObjectMeta.Labels

	pods := &corev1.PodList{}
	options := &client.ListOptions{
		Namespace:     namespace,
		LabelSelector: labels.SelectorFromSet(deployLabels),
	}

	err = kclient.List(context.TODO(), pods, options)
	if err != nil {
		return []corev1.Pod{}, err
	}

	for _, pod := range pods.Items {
		result = append(result, pod)
	}

	return result, err
}
