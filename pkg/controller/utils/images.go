package utils

const (
	// GitlabEnterpriseImage represents the gitlab enterprise edition
	// Image to be deployed in our environment
	GitlabEnterpriseImage = "gitlab/gitlab-ee:12.8.6-ee.0"
	// GitlabCommunityImage represents the gitlab  Community
	// edition image to be deployed
	GitlabCommunityImage = "gitlab/gitlab-ce:12.8.6-ce.0"

	// GitlabRunnerImage represents the runner image
	GitlabRunnerImage = "registry.gitlab.com/ochienged/gitlab-operator/gitlab-runner"

	// GitlabRunnerHelperImage represents the runner image
	GitlabRunnerHelperImage = "registry.gitlab.com/ochienged/gitlab-operator/gitlab-runner-helper"

	// GitalyImage provides high-level RPC access top Git repositories
	GitalyImage = "registry.gitlab.com/gitlab-org/build/cng/gitaly:v12.8.7"

	// GitLabRegistryImage powers the GitLab container registry
	GitLabRegistryImage = "registry.gitlab.com/gitlab-org/build/cng/gitlab-container-registry:v2.8.1-gitlab"

	// GitLabShellImage handless git SSH sessions for GitLab
	GitLabShellImage = "registry.gitlab.com/gitlab-org/build/cng/gitlab-shell:v11.0.0"

	// GitLabWorkhorseImage deploys workhorse which helps alleviate workload from unicorn
	GitLabWorkhorseImage = "registry.gitlab.com/gitlab-org/build/cng/gitlab-workhorse-ee:v12.8.7"

	// GitLabUnicornImage used for pre-forking the Ruby web server to handle requests
	GitLabUnicornImage = "registry.gitlab.com/gitlab-org/build/cng/gitlab-unicorn-ee:v12.8.7"

	// GitLabSidekigImage provides means to deploy a background job processor for GitLab
	GitLabSidekigImage = "registry.gitlab.com/gitlab-org/build/cng/gitlab-sidekiq-ee:v12.8.7"

	// GitLabRunnerImage is for the runners
	GitLabRunnerImage = "gitlab/gitlab-runner:alpine-v12.8.0"

	// GitLabExporterImage is for exporter deployment
	GitLabExporterImage = "registry.gitlab.com/gitlab-org/build/cng/gitlab-exporter:6.0.0"

	// GitLabTaskRunnerImage is for task runner deployment
	GitLabTaskRunnerImage = "registry.gitlab.com/gitlab-org/build/cng/gitlab-task-runner-ee:v12.8.7"

	// RedisExporterImage exports redis metrics for prometheus
	RedisExporterImage = "docker.io/bitnami/redis-exporter:1.3.5-debian-9-r23"

	// PostgresExporterImage is image for PostgreSQL
	PostgresExporterImage = "docker.io/bitnami/postgres-exporter:0.7.0-debian-9-r12"

	// ConfigMapReloadImage provides util to detect configmap changes
	ConfigMapReloadImage = "jimmidyson/configmap-reload:v0.3.0"

	// BusyboxImage used for init containers
	BusyboxImage = "busybox:latest"

	// GitLabCertificatesImage is image for certificates
	GitLabCertificatesImage = "registry.gitlab.com/gitlab-org/build/cng/alpine-certificates:20171114-r3"
)
