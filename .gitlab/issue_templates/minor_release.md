## Summary

This issue will serve as a checklist and record of the first release of the GitLab Operator `X.Y.Z`. Some processes have not been fully automated yet (#224), so this should be helpful to coordinate tasks between the team members and identify points for future improvement.

## To do

1. [ ] Confirm MR is opened automatically by Charts pipeline to update `CHART_VERSIONS` (charts are published around 13:30 UTC) -> https://gitlab.com/gitlab-org/cloud-native/gitlab-operator/-/merge_requests/`NNN`
1. [ ] Confirm MR pipeline passes (ensuring that all new chart versions work as expected) -> https://gitlab.com/gitlab-org/cloud-native/gitlab-operator/-/pipelines/`NNNNNN`
1. [ ] Merge MR to `master`
1. [ ] Create `X-Y-stable` branch from `master` and push
1. [ ] Create release candidate tag `X.Y.0-rc` from `X-Y-stable` and push
1. [ ] Test the release candidate in OpenShift and Kubernetes
1. [ ] Create `X.Y.0` tag from `X-Y-stable` branch and push
1. [ ] Confirm that the release is created with the associated manifest artifacts
1. [ ] Confirm that the tagged image is pushed to the container registry
1. [ ] Include the chart versions in the tag message: `Version X.Y.Z - supports GitLab Charts vX.Y.Z, vX.Y.Z, vX.Y.Z`
1. [ ] Delete any unneeded beta releases from testing (https://docs.gitlab.com/ee/api/releases/#delete-a-release)
1. [ ] Update documentation with references to the latest version

   * [ ] Operator repo
   * [ ] Chart repo
   
### Release cleanup script

```shell
#!/bin/bash
# https://docs.gitlab.com/ee/api/releases/#delete-a-release
# usage:
#   ./cleanup.sh x.y.z

TOKEN="<API token>"
PROJECT_ID="18899486"
RELEASE=$1

curl --request DELETE --header "PRIVATE-TOKEN: ${TOKEN}" \
  "https://gitlab.com/api/v4/projects/${PROJECT_ID}/releases/${RELEASE}"
```
