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

## Requirements
The GitLab operator uses native Kubernetes resources to deploy and manage GitLab in the environment. It therefore will presumably run on any container platform that is derived from Kubernetes.

The operator depends on the Prometheus, Nginx Ingress Controller and Cert Manager operators to achieve some of the tasks it provides to its end users.

## Design Decisions

Decisions made during the design of the operator have been compiled into a
[document](doc/design-decisions.md) with background information to provide
reasoning for reaching the decision.

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
    version: 'X.Y.Z'
    values:
      global:
        hosts:
          domain: example.com
```

## OpenShift Cluster Setup

See [doc/openshift-cluster-setup.md](doc/openshift-cluster-setup.md) for instructions on how to create an OpenShift cluster for development or CI.

## More Information

More information on the operator can be found in the [Operator Documentation](https://gitlab.com/gitlab-org/gl-openshift/documentation) repo.
