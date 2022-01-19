## Summary

This issue will serve as a checklist and record of the release of the GitLab Operator `0.X.Y`. Some processes have not been fully automated yet (#224), so this should be helpful to coordinate tasks between the team members and identify points for future improvement.

## To do

1. [ ] Confirm MR is opened automatically by Charts pipeline to update `CHART_VERSIONS` (charts are published around 13:30 UTC) -> https://gitlab.com/gitlab-org/cloud-native/gitlab-operator/-/merge_requests/`NNN`
1. [ ] Confirm MR pipeline passes (ensuring that all new chart versions work as expected) -> https://gitlab.com/gitlab-org/cloud-native/gitlab-operator/-/pipelines/`NNNNNNNNN`
1. [ ] Merge MR to `master`
1. [ ] Pick relevant changes into `X-Y-stable`
1. [ ] Create `X.Y.Z` tag from `X-Y-stable` branch and push
1. [ ] Confirm that the release is created with the associated manifest artifacts
1. [ ] Confirm that the tagged image is pushed to the container registry
1. [ ] Update Documentation with references to the latest version

   * [ ] Operator repo
   * [ ] Chart repo
