# 6. GitLab application versions supported

Date: 2020-11-03

## Status

Accepted

Related [5. Versioning of the operator](0005-versioning-of-the-operator.md)

## Context

We need to establish correlation between the operator version and GitLab application version (and [GitLab Chart](https://gitlab.com/gitlab-org/charts/gitlab) version)

## Decision

Because the [GitLab Chart](https://gitlab.com/gitlab-org/charts/gitlab) is the source
for most of the objects that the Operator manages, the version of the Operator maps to
versions of the GitLab Chart rather than to the GitLab application itself.

The Operator specifies the supported versions of the chart in the
[CHART_VERSIONS file](https://gitlab.com/gitlab-org/cloud-native/gitlab-operator/-/blob/master/CHART_VERSIONS).

The goal is to support the three latest minor versions of the
[GitLab Chart](https://gitlab.com/gitlab-org/charts/gitlab).

## Consequences

When new versions of the GitLab Chart are released, the CHART_VERSIONS file
will need to be updated. Then, testing of the Operator must be done to ensure
changes in functionality are reflected appropriately in the Operator logic.
