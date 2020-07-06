package gitlab

const (
	// GitalyImage provides high-level RPC access top Git repositories
	GitalyImage = "registry.gitlab.com/gitlab-org/build/cng/gitaly:v13.0.3"

	// GitLabRegistryImage powers the GitLab container registry
	GitLabRegistryImage = "registry.gitlab.com/gitlab-org/build/cng/gitlab-container-registry:v2.9.1-gitlab"

	// GitLabShellImage handless git SSH sessions for GitLab
	GitLabShellImage = "registry.gitlab.com/gitlab-org/build/cng/gitlab-shell:v13.2.0"

	// GitLabWorkhorseImage deploys workhorse which helps alleviate workload from webservice
	GitLabWorkhorseImage = "registry.gitlab.com/gitlab-org/build/cng/gitlab-workhorse-ee:v13.0.3"

	// GitLabWebServiceImage used for pre-forking the Ruby web server to handle requests
	GitLabWebServiceImage = "registry.gitlab.com/gitlab-org/build/cng/gitlab-webservice-ee:v13.0.3"

	// GitLabSidekiqImage provides means to deploy a background job processor for GitLab
	GitLabSidekiqImage = "registry.gitlab.com/gitlab-org/build/cng/gitlab-sidekiq-ee:v13.0.3"

	// GitLabExporterImage is for exporter deployment
	GitLabExporterImage = "registry.gitlab.com/gitlab-org/build/cng/gitlab-exporter:7.0.3"

	// GitLabTaskRunnerImage is for task runner deployment
	GitLabTaskRunnerImage = "registry.gitlab.com/gitlab-org/build/cng/gitlab-task-runner-ee:v13.0.3"

	// RedisImage contains the Redis image
	RedisImage = "docker.io/bitnami/redis:5.0.7-debian-9-r50"

	// RedisExporterImage exports redis metrics for prometheus
	RedisExporterImage = "docker.io/bitnami/redis-exporter:1.3.5-debian-9-r23"

	// BusyboxImage used for init containers
	BusyboxImage = "busybox:latest"

	// GitLabCertificatesImage is image for certificates
	GitLabCertificatesImage = "registry.gitlab.com/gitlab-org/build/cng/alpine-certificates:20171114-r3"

	// PostgresImage is the image for PostgreSQL database
	PostgresImage = "docker.io/bitnami/postgresql:11.7.0"

	// PostgresExporterImage is the image for the Postgres Metrics exporter container
	PostgresExporterImage = "docker.io/bitnami/postgres-exporter:0.8.0-debian-10-r99"

	// MiniDebImage is the image used by the postgres init container to fix filesystem permissions
	MiniDebImage = "docker.io/bitnami/minideb:stretch"

	// MinioImage used to deploy minio cluster
	MinioImage = "minio/minio:RELEASE.2020-01-03T19-12-21Z"

	// MinioClientImage provides the minio client image to use
	MinioClientImage = "minio/mc:RELEASE.2018-07-13T00-53-22Z"
)
