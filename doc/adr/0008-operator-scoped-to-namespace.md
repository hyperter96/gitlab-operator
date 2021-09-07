# 8. Operator scoped to namespace by default

Date: 2020-08-03
Updated: 2021-09-07

## Status

Accepted

## Context

Operator RBAC roles are currently scoped to the entire cluster, but true cluster-scoped support is not ready yet. For example, attempting
to deploy a CR to a namespace other than the Operator will not work because the GitLab service accounts do not exist in that
namespace.

We've seen failures in CI (https://gitlab.com/gitlab-org/cloud-native/gitlab-operator/-/issues/209) related to Operators attempting
to reconcile objects in namespaces other than their own. This is particularly undesirable in CI where each namespace is intended
to be isolated from other releases.

## Decision

Ultimately we want the Operator to be able to resolve GitLab CRs globally, regardless of namespace. Users should easily be able to toggle between global and namespaced behavior.

The Operator supports cluster-scope, but will be scoped to resolve only GitLab CRs deployed to the same the Operator's namespace by default.

Cluster-scope can be configured by disabling the WATCH_NAMESPACE patch that we provide. This is currently experimental.

## Consequences

The Operator is namespace-scoped by default. CI pipelines will pass more reliably as this change ensures that Operators only reconcile objects within their own namespace.

We are investigating global-scope problems in ISSUELINKTODO with the intention of allowing the Operator to resolve GitLab objects in any namespace. We will update this ADR upon resolving that issue.
