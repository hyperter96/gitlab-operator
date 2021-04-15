## Known limitations

Below are known limitations of the GitLab Operator.

### Object storage must use in-cluster MinIO

Currently, the Operator deploys an in-cluster instance of MinIO. This instance must be used for object storage. External object storage providers are not supported at this time.

Related: [#137](https://gitlab.com/gitlab-org/gl-openshift/gitlab-operator/-/issues/137)

### Multiple instances of Webservice, Sidekiq, or Gitaly are not supported

In the GitLab Helm chart, multiple instances of Webservice, Sidekiq, and Gitaly are supported.

The Operator only expects one instance of these components at this time.

Related: [#128](https://gitlab.com/gitlab-org/gl-openshift/gitlab-operator/-/issues/128)

### Installation assumes an OpenShift environment

Portions of the documenation, scripts, and code assume an OpenShift environment.
The plan is for the GitLab Operator to be supported on both OpenShift and "vanilla"
Kubernetes environments.

Progress toward proper support for "vanilla" Kubernetes environments can be tracked
in [#119](https://gitlab.com/gitlab-org/gl-openshift/gitlab-operator/-/issues/119).

### Certain components not supported

Below is a list of unsupported components:

- GitLab Shell [#58](https://gitlab.com/gitlab-org/gl-openshift/gitlab-operator/-/issues/58) is unable to provide SSH access.
- Praefect: [#136](https://gitlab.com/gitlab-org/gl-openshift/gitlab-operator/-/issues/136)
- Pages: [#138](https://gitlab.com/gitlab-org/gl-openshift/gitlab-operator/-/issues/138)
- KAS: [#139](https://gitlab.com/gitlab-org/gl-openshift/gitlab-operator/-/issues/139)
- Mailroom: [#140](https://gitlab.com/gitlab-org/gl-openshift/gitlab-operator/-/issues/140)
