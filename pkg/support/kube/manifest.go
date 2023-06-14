package kube

import (
	"context"
	"strings"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/support/objects"
)

// OwnerReferenceFilter is a context-free callback function that examines an
// OwnerReference object and decides whether to accept it or not.
type OwnerReferenceFilter = func(metav1.OwnerReference) bool

// ManagedObjectDiscoveryConfig is the configuration that is used for finding
// the objects that are managed by a controller.
//
// The configuration requires a Kubernetes API Client to work properly. It also
// requires the list of GroupVersionResources.
//
// You can use different options to provide the Client interface via
// the Manager. Use the available options to configure the runtime behavior of
// DiscoverManagedObjects.
type ManagedObjectDiscoveryConfig struct {
	Client                client.Client
	Context               context.Context
	GroupVersionResources []schema.GroupVersionResource
	Logger                logr.Logger

	owner client.Object
}

// ManagedObjectDiscoveryOption represents an individual option of
// ManagedObjectDiscoveryOption. The available options are:
//
//   - WithClient
//   - WithContext
//   - WithGroupVersionResource
//   - WithGroupVersionResourceArgs
//   - WithLogger
//   - WithManager
//
// See each option for further details.
type ManagedObjectDiscoveryOption = func(*ManagedObjectDiscoveryConfig)

// DiscoverManagedObjects finds all Kubernetes objects that are managed by the
// specified owner. The owner generally is a CustomResource that is reconciled
// by a controller.
//
// One common usage is for a controller to discover its managed resources.
//
// The current implementation relies on the OwnerReferences of objects to
// determine if it is managed by another object. It selects the objects that
// refer to the specified owner. It only uses the APIVersion, Kind, Name, and
// the Namespace of the owner to make the association.
//
// Note that it ignores cluster-scoped resources and only looks up the objects
// that are in the namespace of the owner.
//
// This function returns the list of API resource types that are provided with the
// configuration (for details see WithGroupVersionResource and
// WithGroupVersionResourceArgs). It only looks up the provided resource types
// and ignores the others.
//
// Note that this function uses PartialObjectMetadata which only contains the
// resource metadata. It does not retrieve the managed resources in full, hence
// it can not unmarshal the actual resources into the objects that it returns.
func DiscoverManagedObjects(owner client.Object, options ...ManagedObjectDiscoveryOption) (objects.Collection, error) {
	cfg := defaultManagedObjectDiscoveryConfig(owner)
	cfg.applyOptions(options)

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg.findByOwnerReference()
}

/* ObjectDiscoveryConfig */

func defaultManagedObjectDiscoveryConfig(owner client.Object) *ManagedObjectDiscoveryConfig {
	return &ManagedObjectDiscoveryConfig{
		Context:               context.Background(),
		GroupVersionResources: []schema.GroupVersionResource{},
		Logger:                logr.Discard(),

		owner: owner,
	}
}

func (c *ManagedObjectDiscoveryConfig) applyOptions(options []ManagedObjectDiscoveryOption) {
	for _, option := range options {
		option(c)
	}

	c.Logger = c.Logger.
		WithName("Kube").
		WithValues(
			"name", client.ObjectKeyFromObject(c.owner),
			"kind", c.owner.GetObjectKind().GroupVersionKind())
}

func (c *ManagedObjectDiscoveryConfig) validate() error {
	if c.Client == nil {
		return errors.New("client is required")
	}

	if len(c.GroupVersionResources) == 0 {
		return errors.New("list of resources is required")
	}

	return nil
}

func (c *ManagedObjectDiscoveryConfig) findByOwnerReference() (objects.Collection, error) {
	result := objects.Collection{}

	for _, gvr := range c.GroupVersionResources {
		gvk, _ := c.Client.RESTMapper().KindFor(gvr)

		c.Logger.V(2).Info("fetching resources", "resource", gvr, "kind", gvk.Kind)

		l := &metav1.PartialObjectMetadataList{}
		l.SetGroupVersionKind(gvk)

		if !strings.HasSuffix(l.Kind, "List") {
			l.Kind = l.Kind + "List"
		}

		if err := c.Client.List(c.Context, l, client.InNamespace(c.owner.GetNamespace())); err != nil {
			c.Logger.Error(err, "could not list resources", "resource", gvr)
		} else {
			c.Logger.V(2).Info("got list items", "count", len(l.Items))
		}

		for _, o := range l.Items {
			func(item metav1.PartialObjectMetadata) {
				if item.APIVersion == "" && item.Kind == "" && !gvk.Empty() {
					item.APIVersion = gvk.GroupVersion().String()
					item.Kind = gvk.Kind
				}

				if c.examineOwnerReference(&item) {
					result.Append(&item)
				}
			}(o)
		}
	}

	return result, nil
}

func (c *ManagedObjectDiscoveryConfig) examineOwnerReference(item *metav1.PartialObjectMetadata) bool {
	/* This is just a safeguard. It shouldn't really happen. */
	if item.GetNamespace() != c.owner.GetNamespace() {
		return false
	}

	ownerGVK := c.owner.GetObjectKind().GroupVersionKind()

	for _, ownerRef := range item.OwnerReferences {
		refGV, err := schema.ParseGroupVersion(ownerRef.APIVersion)
		if err != nil {
			c.Logger.Error(err, "WARNING: could not parse owner apiVersion", "apiVersion", ownerRef.APIVersion)
		}

		refGVK := refGV.WithKind(ownerRef.Kind)

		if ownerGVK.Group == refGVK.Group && ownerGVK.Version == refGVK.Version && ownerGVK.Kind == refGVK.Kind {
			if ownerRef.Name == c.owner.GetName() {
				return true
			}
		}
	}

	return false
}
