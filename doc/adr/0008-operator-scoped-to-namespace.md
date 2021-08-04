# 8. Operator scoped to namespace by default

Date: 2020-08-03

## Status

Accepted

## Context

Operator currently is scoped to the entire cluster, but true cluster-scoped support is not ready yet. For example, attempting
to deploy a CR to a namespace other than the Operator will not work because the GitLab service accounts do not exist in that
namespace.

We've seen failures in CI (https://gitlab.com/gitlab-org/cloud-native/gitlab-operator/-/issues/209) related to Operators attempting
to reconcile objects in namespaces other than their own. This is particularly undesirable in CI where each namespace is intended
to be isolated from other releases.

## Decision

The Operator supports cluster-scope, but will be scoped to the namespace by default.
Cluster-scope can be configured by disabling the WATCH_NAMESPACE patch that we provide.

## Consequences

The Operator is namespace-scoped by default. Enabling cluster-scope is done manually until a longer-term decision is made in
https://gitlab.com/gitlab-org/cloud-native/gitlab-operator/-/issues/210.

Also, CI pipelines will pass more reliably as this change ensures that Operators only reconcile objects within their own namespace.
