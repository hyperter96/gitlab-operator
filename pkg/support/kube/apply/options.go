package apply

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/support/kube"
)

/* Apply Options */

// WithClient specifies the Kubernetes API Client that is uses for running
// all cluster operations for ApplyObject.
//
// Note that ApplyObject requires a client and tries different configuration
// points to locate it. However, it does not fall back to a default client.
func WithClient(client client.Client) kube.ApplyOption {
	return func(cfg *kube.ApplyConfig) {
		cfg.Client = client
	}
}

// WithCodec configures apply with the specified codec for serializing and
// deserializing objects.
//
// By default it uses the `UnstructuredJSONScheme` codec.
func WithCodec(codec runtime.Codec) kube.ApplyOption {
	return func(cfg *kube.ApplyConfig) {
		cfg.Codec = codec
	}
}

// WithContext configures apply with the logger and the client from the context.
func WithContext(ctx context.Context) kube.ApplyOption {
	return func(cfg *kube.ApplyConfig) {
		cfg.Context = ctx

		if cfg.Client == nil {
			cfg.Client = getClientFromContext(cfg.Context)
		}

		cfg.Logger = logr.FromContextOrDiscard(ctx)
	}
}

// WithLogger configures apply with the specified logger.
//
// By defaults all logs are discarded.
func WithLogger(logger logr.Logger) kube.ApplyOption {
	return func(cfg *kube.ApplyConfig) {
		cfg.Logger = logger
	}
}

// WithScheme configures apply with the specified scheme for looking up Go types
// from resource kind and API version.
//
// By defaults the global Scheme is used.
func WithScheme(scheme *runtime.Scheme) kube.ApplyOption {
	return func(cfg *kube.ApplyConfig) {
		cfg.Scheme = scheme
	}
}

// WithContext configures apply with the client, scheme, and logger from the
// manager.
//
// This is a convenient way for configuring apply.
func WithManager(manager manager.Manager) kube.ApplyOption {
	return func(cfg *kube.ApplyConfig) {
		cfg.Client = manager.GetClient()
		cfg.Scheme = manager.GetScheme()
		cfg.Logger = manager.GetLogger()
	}
}

/* Private */

func getClientFromContext(ctx context.Context) client.Client {
	/* This is a placeholder for future implementation */
	return nil
}
