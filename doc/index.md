# Operator Documentation

## Installation

Instructions on how to install the GitLab Operator can be found our [installation document](installation.md).

We list [known limitations](#known-limitations), and details of how we use
[Security Context Constraints](security_context_constraints.md) in their respective documents.

You should also be aware of the [considerations for SSH access to Git](git_over_ssh.md), especially
when using OpenShift.

## Upgrading

[Operator upgrades](operator_upgrades.md) documentation demonstrates how to upgrade the GitLab Operator.

[GitLab upgrades](gitlab_upgrades.md) documentation demonstrates how to upgrade a GitLab instance, managed by the GitLab Operator.

## Backup and restore

[Backup and restore](backup_and_restore.md) documentation demonstrates how to back up and restore a GitLab instance that is managed by the Operator.

## Developer Tooling

- [Developer guide](developer/guide.md): Outlines the project structure and how to contribute.
- [Versioning and Release Info](developer/releases.md): Records notes concerning versioning and releasing the operator.
- [Design decisions](https://gitlab.com/gitlab-org/cloud-native/gitlab-operator/-/blob/master/doc/adr): This projects makes use of Architecture Descision Records, detailing the structure, functionality, and feature implementation of this Operator.
- [OpenShift Cluster Setup](developer/openshift_cluster_setup.md): Instructions for creating/configuring OpenShift clusters for *Development* purposes.

## Known limitations

Below are known limitations of the GitLab Operator.

### Components not supported

Below is a list of unsupported components:

- Praefect: [#136](https://gitlab.com/gitlab-org/cloud-native/gitlab-operator/-/issues/136)

### Components not recommended for production use

Below is a list of components that are not yet recommended for production when
deployed by the GitLab Operator:

- KAS: [#353](https://gitlab.com/gitlab-org/cloud-native/gitlab-operator/-/issues/353)

### Components not sourced from GitLab Helm Charts

- MinIO: [#374](https://gitlab.com/gitlab-org/cloud-native/gitlab-operator/-/issues/374)