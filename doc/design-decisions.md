# Design Decisions

This document provides a place to collect and document decisions made
regarding the design of the GitLab OpenShift operator.

## Integration of the GitLab Chart

The Operator renders the [GitLab Chart](https://gitlab.com/gitlab-org/charts/gitlab) using the `values`
field from the GitLab CR to create a template similar to the output of `helm template`.

The Operator queries this template for objects to deploy based on the configuration provided in the
CR values.

Leveraging the GitLab Chart greatly accelerates the progression of the GitLab Operator by capturing the
objects and logic from the charts. This means the Operator is effectively a wrapper around the GitLab Chart,
with additional capabilities aiming to fulfill the
[Operator maturity model](https://docs.openshift.com/container-platform/4.1/applications/operators/olm-what-operators-are.html#olm-maturity-model_olm-what-operators-are).

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

The version of the operator shall follow semantic versioning during development
until the feature set is sufficient to support the GitLab application
lifecycle through the operator. At which point the version will be bumped
to be in sync with the GitLab application version.

Further information can be found in https://gitlab.com/gitlab-org/gl-openshift/gitlab-operator/-/issues/5

## GitLab application versions supported

Because the [GitLab Chart](https://gitlab.com/gitlab-org/charts/gitlab) is the source
for most of the objects that the Operator manages, the version of the Operator maps to
versions of the GitLab Chart rather than to the GitLab application itself.

The Operator specifies the supported versions of the chart in the
[CHART_VERSIONS file](../CHART_VERSIONS).

The goal is to support the three latest minor versions of the
[GitLab Chart](https://gitlab.com/gitlab-org/charts/gitlab).

When new versions of the GitLab Chart are released, the CHART_VERSIONS file
will need to be updated. Then, testing of the Operator must be done to ensure
changes in functionality are reflected appropriately in the Operator logic.

## Dependency on other external components

The Helm charts include charts for `cert-manager` and `nginx-ingress`.
In contrast, the Operator does not deploy these components and therefore
requires them to be installed separately.

This keeps the codebase focused on GitLab functionality, and ensures that
the Operator is not tied to specific certificates or ingress managers.
