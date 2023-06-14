package manifest

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/support/kube"
)

/* DiscoverManagedObjects Options */

// WithClient specifies the Kubernetes API Client that is used for listing
// object metadata.
//
// The client is required for DiscoverManagedObjects and you must provide it
// either directly with WithClient or use WithManager.
func WithClient(client client.Client) kube.ManagedObjectDiscoveryOption {
	return func(cfg *kube.ManagedObjectDiscoveryConfig) {
		cfg.Client = client
	}
}

// WithContext configures DiscoverManagedObjects with the specified context and
// its associated logger if there is any.
//
// By default it uses the background context.
func WithContext(ctx context.Context) kube.ManagedObjectDiscoveryOption {
	return func(cfg *kube.ManagedObjectDiscoveryConfig) {
		cfg.Context = ctx
		cfg.Logger = logr.FromContextOrDiscard(ctx)
	}
}

// WithGroupVersionResources configures DiscoverManagedObjects with the
// specified Kubernetes API resource types. This is required for the default
// mode but is ignored in auto-discovery mode.
//
// See WithGroupVersionResourceArgs for a more convenient way to provide this
// list.
func WithGroupVersionResources(resources ...schema.GroupVersionResource) kube.ManagedObjectDiscoveryOption {
	return func(cfg *kube.ManagedObjectDiscoveryConfig) {
		cfg.GroupVersionResources = resources
	}
}

// WithGroupVersionResourceArgs parses the provided string arguments into a list
// of Kubernetes API resource types and configures DiscoverManagedObjects to
// use it.
//
// Each resource argument must present a GroupVersionResource in
// `resource.version.group.com` format. Note that the version must be included
// because it does not infer the preferred server version.
func WithGroupVersionResourceArgs(resources ...string) kube.ManagedObjectDiscoveryOption {
	return func(cfg *kube.ManagedObjectDiscoveryConfig) {
		grvs := make([]schema.GroupVersionResource, 0, len(resources))

		for _, r := range resources {
			grv, _ := schema.ParseResourceArg(r)
			if grv != nil {
				grvs = append(grvs, *grv)
			}
		}

		cfg.GroupVersionResources = grvs
	}
}

// WithLogger configures DiscoverManagedObjects with the specified logger.
//
// By defaults all logs are discarded.
func WithLogger(logger logr.Logger) kube.ManagedObjectDiscoveryOption {
	return func(cfg *kube.ManagedObjectDiscoveryConfig) {
		cfg.Logger = logger
	}
}

// WithManager is a convenient option that retrieves multiple configuration
// items from the manager, including the client and the logger.
func WithManager(manager manager.Manager) kube.ManagedObjectDiscoveryOption {
	return func(cfg *kube.ManagedObjectDiscoveryConfig) {
		cfg.Client = manager.GetClient()
		cfg.Logger = manager.GetLogger()
	}
}
