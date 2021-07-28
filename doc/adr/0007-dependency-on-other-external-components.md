# 7. Dependency on other external components

Date: 2020-12-01

## Status

Accepted

## Context

Operator depends on external components. We need to define how are we going to approach such external dependencies moving forward.

## Decision

The Helm charts include a chart for `cert-manager`.
In contrast, the Operator does not deploy this component and therefore
requires it to be installed separately.

## Consequences

This keeps the codebase focused on GitLab functionality, and ensures that
the Operator is not tied to specific certificate managers.