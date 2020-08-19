package gitlab

import (
	"fmt"
	"io/ioutil"
	"strings"

	"path/filepath"

	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/pkg/apis/gitlab/v1beta1"
	"sigs.k8s.io/yaml"
)

// Image represents a
// single microservice image
type Image struct {
	Name  string `yaml:"name"`
	Image string `yaml:"image"`
}

// Release defines a GitLab release
type Release struct {
	Version string  `json:"version" yaml:"version"`
	Default bool    `json:"default,omitempty" yaml:"default,omitempty"`
	Images  []Image `json:"images" yaml:"images"`
}

// DefaultRelease returns release marked as default
func DefaultRelease() *Release {
	releases := getReleaseList()

	for _, release := range releases {
		if release.IsDefault() {
			return release
		}
	}

	log.Error(fmt.Errorf("No release is marked as default"), "Release was not found")

	return nil
}

// BuildRelease returns user specified release
// returns default release if custom release is not found
func BuildRelease(cr *gitlabv1beta1.Gitlab) *Release {
	release := getRelease(fmt.Sprintf("release_%s.yaml", cr.Spec.Release))
	if release != nil {
		return release
	}

	return DefaultRelease()
}

func getImage(release *Release, target string) string {
	for _, container := range release.Images {
		if container.Name == target {
			return container.Image
		}
	}
	return ""
}

// getRelease gets the release filename
// and returns a release object
func getRelease(filename string) *Release {
	releaseFile, err := ioutil.ReadFile(filepath.Join("releases", filename))
	if err != nil {
		return nil
	}

	release := &Release{}
	err = yaml.Unmarshal(releaseFile, release)
	if err != nil {
		return nil
	}

	return release
}

// returns a list of releases in releases directory
func getReleaseList() []*Release {
	var releases []*Release
	contents, err := ioutil.ReadDir("/releases/")
	if err != nil {
		log.Error(err, "Error listing directory contents")
		return nil
	}

	for _, releaseFile := range contents {
		if strings.HasPrefix(releaseFile.Name(), "release_") {
			releases = append(releases, getRelease(releaseFile.Name()))
		}
	}

	return releases
}

// IsDefault returns true if release is set to default
func (r *Release) IsDefault() bool {
	return r != nil && r.Default == true
}

// Gitaly return gitaly container image
func (r *Release) Gitaly() string {
	return getImage(r, "gitaly")
}

// Sidekiq return sidekiq container image
func (r *Release) Sidekiq() string {
	return getImage(r, "sidekiq")
}

// Workhorse return workhorse container image
func (r *Release) Workhorse() string {
	return getImage(r, "workhorse")
}

// Webservice return webservice container image
func (r *Release) Webservice() string {
	return getImage(r, "webservice")
}

// Registry return registry container image
func (r *Release) Registry() string {
	return getImage(r, "registry")
}

// Shell return shell container image
func (r *Release) Shell() string {
	return getImage(r, "shell")
}

// TaskRunner return task runner container image
func (r *Release) TaskRunner() string {
	return getImage(r, "task_runner")
}

// GitLabExporter return GitLab exporter container image
func (r *Release) GitLabExporter() string {
	return getImage(r, "gitlab_exporter")
}

// Redis return redis container image
func (r *Release) Redis() string {
	return getImage(r, "redis")
}

// RedisExporter return redis exporter container image
func (r *Release) RedisExporter() string {
	return getImage(r, "redis_exporter")
}

// Postgresql return postgresql container image
func (r *Release) Postgresql() string {
	return getImage(r, "postgresql")
}

// PostgresqlExporter return postgres exporter container image
func (r *Release) PostgresqlExporter() string {
	return getImage(r, "postgresql_exporter")
}

// Minio return minio container image
func (r *Release) Minio() string {
	return getImage(r, "minio")
}

// MinioClient return minio client container image
func (r *Release) MinioClient() string {
	return getImage(r, "minio_client")
}

// Busybox return busybox container image
func (r *Release) Busybox() string {
	return getImage(r, "busybox")
}

// Certificates return certificates container image
func (r *Release) Certificates() string {
	return getImage(r, "certificates")
}

// MiniDebian return mini debian container image
func (r *Release) MiniDebian() string {
	return getImage(r, "mini_deb")
}
