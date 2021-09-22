# Operator Documentation

## Installation

Instructions on how to install the GitLab Operator can be found our [installation document](installation.md).

We list [known limitations](known-limitations.md), and details of how we use
[Security Context Constraints](security-context-constraints.md) in their respective documents.

You should also be aware of the [considerations for SSH access to Git](git-over-ssh.md), especially
when using OpenShift.

## Upgrading

[Operator upgrades](operator-upgrades.md) documentation demonstrates how to upgrade the GitLab Operator.

[GitLab upgrades](gitlab-upgrades.md) documentation demonstrates how to upgrade a GitLab instance, managed by the GitLab Operator.

## Backup and restore

[Backup and restore](backup-and-restore.md) documentation demonstrates how to back up and restore a GitLab instance that is managed by the Operator.

## Developer Tooling

- [Developer guide](developer/guide.md): Outlines the project structure and how to contribute.
- [Versioning and Release Info](developer/releases.md): Records notes concerning versioning and releasing the operator.
- [Design decisions](adr/): This projects makes use of Architecture Descision Records, detailing the structure, functionality, and feature implementation of this Operator.
- [OpenShift Cluster Setup](openshift-cluster-setup.md): Instructions for creating/configuring OpenShift clusters for *Development* purposes.
