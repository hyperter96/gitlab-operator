# 14. Supported OpenShift versions

Date: 2021-11-24

## Status

Proposed

## Context

We can realistically support a few versions of OpenShift, so we need
a rolling schedule of which versions we currently support. We use a similar
rolling schedule for GitLab charts and supported Kubernetes versions. For
charts we based our support schedule on the versions of managed Kubernetes popular
cloud providers offer.

Managed OpenShift offerings from
[Red Hat](https://docs.openshift.com/dedicated/osd_policy/osd-life-cycle.html#rosa-life-cycle-dates_osd-life-cycle),
[Azure](https://docs.microsoft.com/en-us/azure/openshift/support-lifecycle#azure-red-hat-openshift-release-calendar), and
[AWS](https://cloud.redhat.com/products/amazon-openshift) were consulted.
[GCP](https://cloud.google.com/architecture/partners/openshift-on-gcp) does not offer a managed OpenShift service.

## Decision

Since Azure managed OpenShift lags behind releases for three months, we will test and support `N - 2` minor versions of OpenShift at a time, removing testing clusters three months after official Red Hat support ends. This will ensure that we allow users on Azure enough time to upgrade their managed OpenShift clusters.

For our testing/support that means we should currently have 4.7, 4.8 and 4.9 clusters.
Once OpenShift 4.10 (or 5.0) is released, we should plan to remove 4.7 from our projects three months after Red Hat EOL, February 2022.

This cycle corresponds to supporting each minor version of OpenShift for one year after release.

## Consequences

GitLab Operator users will have to be made aware that we officially support `N-2` versions
of OpenShift. We may have to issue deprecation warnings from the operator when running
in an OpenShift cluster version that has aged out of support.
