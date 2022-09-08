package adapter

import (
	"context"

	apiv1beta1 "gitlab.com/gitlab-org/cloud-native/gitlab-operator/api/v1beta1"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/gitlab"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/gitlab/internal/v1beta1"
)

// NewV1Beta1 creates a new wrapper for the specified GitLab resource.
func NewV1Beta1(ctx context.Context, src *apiv1beta1.GitLab) (gitlab.Adapter, error) {
	return v1beta1.NewAdapter(ctx, src)
}
