package internal

import (
	"context"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"time"

	"crypto/sha256"
	"encoding/json"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// Label function returns uniform labels for resources.
func Label(resource, component, resourceType string) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":       resource,
		"app.kubernetes.io/instance":   strings.Join([]string{resource, component}, "-"),
		"app.kubernetes.io/component":  component,
		"app.kubernetes.io/part-of":    resourceType,
		"app.kubernetes.io/managed-by": "gitlab-operator",
	}
}

// ResourceQuantity returns quantity requested for resource.
func ResourceQuantity(request string) resource.Quantity {
	return resource.MustParse(request)
}

// IsGroupVersionSupported checks for API endpoint for given Group and Version.
func IsGroupVersionSupported(group, version string) bool {
	// For unit tests, we need to skip client creation.
	if os.Getenv("USE_EXISTING_CLUSTER") == "false" {
		return false
	}

	client, err := KubernetesConfig().NewKubernetesClient()
	if err != nil {
		fmt.Printf("Unable to acquire k8s client: %v", err)
	}

	groupVersion := schema.GroupVersion{
		Group:   group,
		Version: version,
	}

	if err := discovery.ServerSupportsVersion(client, groupVersion); err != nil {
		return false
	}

	return true
}

// IsOpenshift check if API has the API route.openshift.io/v1,
// then it is considered an openshift environment.
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

// KubeConfig returns kubernetes client configuration.
type KubeConfig struct {
	Config *rest.Config
	Error  error
}

// KubernetesConfig returns kubernetes client config.
func KubernetesConfig() KubeConfig {
	config, err := config.GetConfig()
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

// NewKubernetesClient returns a client that can be used to interact with the kubernetes api.
func (k KubeConfig) NewKubernetesClient() (*kubernetes.Clientset, error) {
	conf := k.Config

	if err := k.Error; err != nil {
		panic(fmt.Sprintf("Error getting cluster config: %v", err))
	}

	return kubernetes.NewForConfig(conf)
}

// GetSecretValue returns the value for a key from an existing secret.
func GetSecretValue(client client.Client, namespace, secret, key string) (string, error) {
	target := &corev1.Secret{}

	err := client.Get(context.TODO(), types.NamespacedName{Name: secret, Namespace: namespace}, target)
	if err != nil {
		return "", err
	}

	return string(target.Data[key]), err
}

// ReadConfig returns contents of file as string.
func ReadConfig(filename string) string {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		// log.Error(err, "Error reading %s", filename)
		fmt.Printf("Error reading %s: %v", filename, err)
	}

	return string(content)
}

// Password creates an password string based on the password options provided.
func Password(options PasswordOptions) string {
	password := make([]byte, options.Length)
	charset := []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz01234567890")
	special := []byte("+,#?$()&%^*=/!+<>[]{}@-_")

	// add special characters if not database password
	if options.EnableSpecialChars {
		charset = append(charset, special...)
	}

	//nolint:gosec // This will go away with #137, which will remove Minio objects from the controller.
	seed := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := range password {
		password[i] = charset[seed.Intn(len(charset))]
	}

	return string(password)
}

// SecretData gets a secret by name and returns its data.
func SecretData(name, namespace string) (map[string]string, error) {
	client, err := KubernetesConfig().NewKubernetesClient()
	if err != nil {
		return map[string]string{}, err
	}

	secret, err := client.CoreV1().Secrets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return map[string]string{}, err
	}

	return secret.StringData, nil
}

// ConfigMapData returns data contained in a configmap.
func ConfigMapData(name, namespace string) (map[string]string, error) {
	client, err := KubernetesConfig().NewKubernetesClient()
	if err != nil {
		return map[string]string{}, err
	}

	cm, err := client.CoreV1().ConfigMaps(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return map[string]string{}, err
	}

	return cm.Data, nil
}

// GetDeploymentPods returns the pods that belong to a given deployment.
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

	result = append(result, pods.Items...)

	return result, err
}

// ConfigMapWithHash updates configmap with
// annotation containing a SHA256 hash of its data.
func ConfigMapWithHash(cm *corev1.ConfigMap) {
	jdata, err := json.Marshal(cm.Data)
	if err != nil {
		return
	}

	hash := sha256.Sum256(jdata)

	cm.Annotations = map[string]string{
		"checksum": hex.EncodeToString(hash[:]),
	}
}

// DeploymentConfigMaps returns a list of configmaps used in a deployment.
func DeploymentConfigMaps(deploy *appsv1.Deployment) []string {
	cms := []string{}

	for _, container := range deploy.Spec.Template.Spec.Containers {
		if len(container.Env) != 0 {
			for _, env := range container.Env {
				if env.ValueFrom != nil {
					if env.ValueFrom.ConfigMapKeyRef != nil {
						cms = append(cms, env.ValueFrom.ConfigMapKeyRef.LocalObjectReference.Name)
					}
				}
			}
		}
	}

	for _, vol := range deploy.Spec.Template.Spec.Volumes {
		if vol.VolumeSource.ConfigMap != nil {
			cms = append(cms, vol.VolumeSource.ConfigMap.LocalObjectReference.Name)
		}

		if vol.VolumeSource.Projected != nil {
			for _, source := range vol.VolumeSource.Projected.Sources {
				if source.ConfigMap != nil {
					cms = append(cms, source.ConfigMap.LocalObjectReference.Name)
				}
			}
		}
	}

	return cms
}
