## Summary

This issue will serve as a checklist and record of releases for GitLab Operator `X.Y.Z`.

Much of the process is automated by [release-tools](https://gitlab.com/gitlab-org/release-tools).

## To do

1. [ ] Confirm that stable branch and tag exist in all mirrors of the project:
   * [ ] [Canonical repo](https://gitlab.com/gitlab-org/cloud-native/gitlab-operator)
   * [ ] [Security mirror](https://gitlab.com/gitlab-org/security/cloud-native/gitlab-operator/)
   * [ ] [Build mirror](https://dev.gitlab.org/gitlab/cloud-native/gitlab-operator/)
1. [ ] Confirm that publish jobs in the
       [tag pipeline in Canonical repo](https://gitlab.com/gitlab-org/cloud-native/gitlab-operator/-/pipelines?ref=X.Y.Z)
       has completed.
1. [ ] Confirm that the release is created in the
       [releases page](https://gitlab.com/gitlab-org/cloud-native/gitlab-operator/-/releases)
       with the associated manifest artifacts.
1. [ ] Confirm that the tagged image is pushed to the
       [container registry](https://gitlab.com/gitlab-org/cloud-native/gitlab-operator/container_registry).
1. [ ] [Release to OperatorHub.io](doc/developer/operatorhub_publishing.md).
1. [ ] [Release to OpenShift OperatorHub.io](doc/developer/operatorhub_publishing.md)

/assign @mnielsen @pursultani
