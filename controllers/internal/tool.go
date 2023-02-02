package internal

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// ResourceLabels function returns uniform labels for resources.
func ResourceLabels(resource, component, resourceType string) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":       resource,
		"app.kubernetes.io/instance":   strings.Join([]string{resource, component}, "-"),
		"app.kubernetes.io/component":  component,
		"app.kubernetes.io/part-of":    resourceType,
		"app.kubernetes.io/managed-by": "gitlab-operator",
	}
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
