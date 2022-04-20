## Summary

This issue will serve as a checklist and record of releases for GitLab Operator `X.Y.Z`.

Much of the process is automated by [release-tools](https://gitlab.com/gitlab-org/release-tools).

## To do

1. [ ] Confirm that the stable branch `X-Y-stable` is created and that the associated pipeline passes.
1. [ ] If the stable branch pipeline fails, submit any required changes to fix it.
1. [ ] Confirm that the tag `X.Y.Z` is created and that the associated pipeline passes.
1. [ ] Run the manual `upload_manifest` job in the tag pipeline to release the manifests for the release.
1. [ ] Confirm that the release is created in the
       [releases page](https://gitlab.com/gitlab-org/cloud-native/gitlab-operator/-/releases)
       with the associated manifest artifacts.
1. [ ] Confirm that the tagged image is pushed to the
       [container registry](https://gitlab.com/gitlab-org/cloud-native/gitlab-operator/container_registry).
1. [ ] Update [installation documentation](../../doc/installation.md) with references to the latest version.
1. [ ] [Release to OperatorHub.io](doc/developer/operatorhub_publishing.md).
