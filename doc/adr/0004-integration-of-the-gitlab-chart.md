# 4. Integration of the GitLab Chart

Date: 2020-11-16

## Status

Accepted

## Context

Leveraging the GitLab Chart greatly accelerates the progression of the GitLab Operator by capturing the
objects and logic from the charts. 

## Decision

The Operator will render the [GitLab Chart](https://gitlab.com/gitlab-org/charts/gitlab) using the `values`
field from the GitLab CR to create a template similar to the output of `helm template`.

The Operator will query this template for objects to deploy based on the configuration provided in the
CR values.

## Consequences

This means the Operator is effectively a wrapper around the GitLab Chart,
with additional capabilities aiming to fulfill the
[Operator maturity model](https://docs.openshift.com/container-platform/4.1/applications/operators/olm-what-operators-are.html#olm-maturity-model_olm-what-operators-are).