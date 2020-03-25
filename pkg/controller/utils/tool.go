package utils

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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

// ResourceQuantity returns quantity requested for resource
func ResourceQuantity(request string) resource.Quantity {
	return resource.MustParse(request)
}

// IsPrometheusSupported checks for Prometheus API endpoint
func IsPrometheusSupported() bool {
	client, err := NewKubernetesClient()
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
	client, err := NewKubernetesClient()
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

// NewKubernetesClient returns a client that can be used to interact
// with the kubernetes api
func NewKubernetesClient() (clientset *kubernetes.Clientset, err error) {
	conf, err := rest.InClusterConfig()
	if err != nil {
		fmt.Printf("Unable getting cluster config: %v", err)
	}

	clientset, err = kubernetes.NewForConfig(conf)
	return
}

// GetSecretValue returns the value for a key from an existing secret
func GetSecretValue(client client.Client, namespace, secret, key string) string {
	target := &corev1.Secret{}

	err := client.Get(context.TODO(), types.NamespacedName{Name: secret, Namespace: namespace}, target)
	if err != nil {
		fmt.Printf("Secret not found: %v", err)
	}
	return string(target.Data[key])
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
