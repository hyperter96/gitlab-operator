[![Go Report Card](https://goreportcard.com/badge/gitlab.com/gitlab-org/gl-openshift/gitlab-operator "Go Report Card")](https://goreportcard.com/report/gitlab.com/gitlab-org/gl-openshift/gitlab-operator)

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

[![Go Reference](https://pkg.go.dev/badge/gitlab.com/gitlab-org/gl-openshift/gitlab-operator.svg)](https://pkg.go.dev/gitlab.com/gitlab-org/gl-openshift/gitlab-operator)

# GitLab Operator

The GitLab operator aims to manage the full lifecycle of GitLab instances in your Kubernetes or Openshift container platforms.

While new and still actively being developed, the operator aims to:

- ease installation and configuration of GitLab instances
- offer seamless upgrades from version to version
- ease backup and restore of GitLab and its components
- aggregate and visualize metrics using Prometheus and Grafana
- setup auto-scaling

Note that this does not include the GitLab Runner. For more inforation, see the [GitLab Runner Operator repository](https://gitlab.com/gitlab-org/gl-openshift/gitlab-runner-operator).

## Documentation

More information on the Operator can be found in the [Operator Documentation](doc/README.md).

## Owned Custom Resource: GitLab

The operator is responsible for owning, watching, and reconciling the GitLab custom resource.

An example GitLab object is shown below:

```yaml
apiVersion: apps.gitlab.com/v1beta1
kind: GitLab
metadata:
  name: example
spec:
  chart:
    version: "X.Y.Z"
    values:
      global:
        hosts:
          domain: example.com
```
