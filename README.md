# gitlab-operator
[![Go Report Card](https://goreportcard.com/badge/gitlab.com/ochienged/gitlab-operator "Go Report Card")](https://goreportcard.com/report/gitlab.com/ochienged/gitlab-operator)

The GitLab operator creates and manages GitLab instances/deployments in a container platform such as Openshift or Kubernetes.

## Requirements
The GitLab operator uses native kubernetes resources to deploy and manage GitLab in the environment. It therefore will run an any environment that provides deployments, statefulsets, services, ingress/openshift routes, persistent volume claims, persistent volumes, etc.

## GitLab Operator
The Gitlab Operator introduces two new resource types at this time:  

* [Gitlab](docs/gitlab.md)
* [Runner](docs/runner.md)

The Gitlab resource would deploy the Gitlab application and or the registry depending on the resource spec provided when creating the resource.

The Runner resource would be used to create a resource that can be used to integrate with the

## Clean up
At this time, the operator does not delete the Redis and database (PostgreSQL) data volumes when a GitLab instance is deleted. Therefore, remember to delete any lingering volumes as we look into resolving this issue.

In addition, when you delete or uninstall the operator from the environment, you will need to remove the CRDs.

```
$ kubectl delete crd gitlabs.gitlab.com
$ kubectl delete crd runners.gitlab.com
```
