## Summary

This issue will serve as a checklist and record of the first release of the GitLab Operator `X.Y.Z`. Some processes have not been fully automated yet (#224), so this should be helpful to coordinate tasks between the team members and identify points for future improvement.

## To do

1. [ ] Confirm MR is opened automatically by Charts pipeline to update `CHART_VERSIONS` (charts are published around 13:30 UTC) -> https://gitlab.com/gitlab-org/cloud-native/gitlab-operator/-/merge_requests/`NNN`
1. [ ] Confirm MR pipeline passes (ensuring that all new chart versions work as expected) -> https://gitlab.com/gitlab-org/cloud-native/gitlab-operator/-/pipelines/`NNNNNN`
1. [ ] Merge MR to `master`
1. [ ] Create `X-Y-stable` branch from `master` and push
1. [ ] Create `X.Y.0` tag from `X-Y-stable` branch: `./scripts/tag_release.sh <version>`
1. [ ] Push `X.Y.0` tag: `git push origin <version>`
1. [ ] Confirm that the release is created with the associated manifest artifacts
1. [ ] Confirm that the tagged image is pushed to the container registry
1. [ ] Update documentation with references to the latest version
   * [ ] Operator repo
   * [ ] Chart repo
1. [ ] [Release to OperatorHub.io](doc/developer/operatorhub_publishing.md)
