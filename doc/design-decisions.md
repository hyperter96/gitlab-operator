# Design Decisions

This document provides a place to collect and document decisions made
regarding the design of the GitLab OpenShift operator.

## Versioning of the operator

The intention is to version the operator matching the version number as the
GitLab application and be on the same monthly release schedule. There are
several hurdles to overcome and achieve matching the version numbers.

- Operator needs to support a greater feature parity with GitLab Helm chart
- Come to consensus on single operator or splitting into application and
  Runner operators. Issue: https://gitlab.com/gitlab-org/gl-openshift/gitlab-operator/-/issues/30
- Proper handling of AnyUID in GitLab application containers. Issue:
  https://gitlab.com/gitlab-org/gl-openshift/gitlab-operator/-/issues/11
- Decide on the level of namespace scoping that the operator will be capable of.
  Issue: https://gitlab.com/gitlab-org/gl-openshift/gitlab-operator/-/issues/31

Currently the operator is managing the lifecycle of the GitLab Runner which
is released monthly with the GitLab application. The features being developed
for the GitLab Runner are not significantly impacting the development of the
operator.

The version of the operator shall follow semantic versioning during development
until the feature set is sufficient to support the GitLab application
lifecycle through the operator. At which point the version will be bumped
to be in sync with the GitLab application version.

Further information can be found in https://gitlab.com/gitlab-org/gl-openshift/gitlab-operator/-/issues/5

## GitLab application versions supported

The operator follows the support model of the GitLab application and supports
the latest three releases. This has been done to simply and make the code base
more robust. In addition, pruning the code supporting older releases should
reduce the amount of technical debt incurred in the code base.

## GitLab Runner versions supported

The configuration of the GitLab Runner has a fairly stable format and in
most cases the operator will correctly manage multiple versions of the
Runner. Even so, the Runner Helm chart is pre-1.0 and the deployment
artifacts may change significantly. Given the uncertainty of signifiant
change in the Runner chart the operator supports the latest version of
GitLab Runner. When the Runner chart matures to a 1.0 version the decision
of supporting multiple Runner versions will be revisited.
