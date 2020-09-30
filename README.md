[![Go Report Card](https://goreportcard.com/badge/gitlab.com/gitlab-org/gl-openshift/gitlab-operator "Go Report Card")](https://goreportcard.com/report/gitlab.com/gitlab-org/gl-openshift/gitlab-operator)

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

# GitLab Operator

The GitLab operator aims to manage the full lifecycle of GitLab and GitLab Runner instances in your Kubernetes or Openshift container platforms.

While new and still actively being developed, the operator aims to:

- ease installation and configuration of GitLab and GitLab Runner instances
- offer seamless upgrades from version to version
- ease backup and restore of GitLab and its components
- aggregate and visualize metrics using Prometheus and Grafana
- setup auto-scaling

## Requirements
The GitLab operator uses native Kubernetes resources to deploy and manage GitLab in the environment. It therefore will presumably run on any container platform that is derived from Kubernetes.

The operator depends on the Prometheus, Nginx Ingress Controller and Cert Manager operators to achieve some of the tasks it provides to its end users.

## Owned Resources

The operator is responsible for owns, watches and reconciles three different primary resources at this time.

#### 1. GitLab
GitLab is a complete open-source DevOps platform, delivered as a single application. It fundamentally changes the way development, security and Ops teams collaborate and build software.

An example GitLab object is shown below:

```
apiVersion: apps.gitlab.com/v1beta1
kind: GitLab
metadata:
  name: example
spec:
  url: gitlab.example.com
  volume:
    capacity: 10Gi
  registry:
    url: registry.example.com
  redis:
    volume:
      capacity: 4Gi
  postgres:
    volume:
      capacity: 8Gi
  autoscaling:
    minReplicas: 2
    maxReplicas: 5
    targetCPU: 60

```

#### 2. GitLab Runner
GitLab Runner is the open source project used to run your jobs and send the results back to GitLab. It is used in conjunction with GitLab CI/CD, the open-source continuous integration service included with GitLab that coordinates jobs.

An example is shown below:

```
apiVersion: apps.gitlab.com/v1beta1
kind: Runner
metadata:
  name: example
spec:
  gitlab:
    url: https://gitlab.com
  token: runner-token-secret # Name of the secret containing the Runner token
  tags: openshift, test
```

#### 3. GitLab Backup
The GitLab Backup resource is responsible for performing backups and restores of GitLab instances.

An example is shown below:

```
apiVersion: apps.gitlab.com/v1beta1
kind: GLBackup
metadata:
  name: example
spec:
  instance: example
  schedule: 30 1 1 * 6
  skip: repository,registry
```

## More Information

More information on the operator can be found in the [Operator Documentation](https://gitlab.com/gitlab-org/gl-openshift/documentation) repo.