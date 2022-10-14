package component

import (
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/gitlab"
)

const (
	GitLab gitlab.Component = "gitlab"

	Gitaly         gitlab.Component = "gitaly"
	GitLabExporter gitlab.Component = "gitlab-exporter"
	GitLabPages    gitlab.Component = "gitlab-pages"
	GitLabShell    gitlab.Component = "gitlab-shell"
	GitLabKAS      gitlab.Component = "kas"
	Mailroom       gitlab.Component = "mailroom"
	Migrations     gitlab.Component = "migrations"
	MinIO          gitlab.Component = "minio"
	NginxIngress   gitlab.Component = "nginx-ingress"
	PostgreSQL     gitlab.Component = "postgresql"
	Praefect       gitlab.Component = "praefect"
	Prometheus     gitlab.Component = "prometheus"
	Redis          gitlab.Component = "redis"
	Registry       gitlab.Component = "registry"
	SharedSecrets  gitlab.Component = "shared-secrets"
	Sidekiq        gitlab.Component = "sidekiq"
	Spamcheck      gitlab.Component = "spamcheck"
	Toolbox        gitlab.Component = "toolbox"
	Webservice     gitlab.Component = "webservice"
)

var (
	Core = gitlab.Components{
		PostgreSQL,
		Redis,
		Gitaly,
	}

	Stateful = gitlab.Components{
		PostgreSQL,
		Redis,
		Gitaly,
		MinIO,
	}

	All = gitlab.Components{
		Gitaly,
		GitLabExporter,
		GitLabPages,
		GitLabShell,
		GitLabKAS,
		Mailroom,
		Migrations,
		MinIO,
		NginxIngress,
		PostgreSQL,
		Redis,
		Registry,
		SharedSecrets,
		Sidekiq,
		Webservice,
	}
)
