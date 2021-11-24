# Versioning

The GitLab Operator uses [semver versioning](https://semver.org/).
Version tags should [be the semver version string](../adr/0009-version-tagging.md).

## Documentation

Operator documentation is available in the `doc/` directory.

The [Operator installation document](../../installation.md) is mirrored into the
[GitLab documentation site](https://docs.gitlab.com/charts/installation/operator.html),
primarily for public visibility.

At this time, this process is manual, which means changes to this document requires
two merge requests:

- Operator: [doc/installation.md](../../installation.md)
- Charts: [doc/installation/operator.md](https://gitlab.com/gitlab-org/charts/gitlab/-/blob/master/doc/installation/operator.md)

Issue [#418](https://gitlab.com/gitlab-org/cloud-native/gitlab-operator/-/issues/418)
tracks work toward automating this process.

## Red Hat Certification

The release pipeline will contain a `certification_upload` job when the
repository has been tagged with a semver version
(i.e. `1.0.0`). This job will upload the operator image to Red Hat for
certification test and allow the operator to be published through the
Red Hat Connect portal.

It is also possible to pass a release candidate tag (i.e. `1.0.0-rc1`) or a
beta tag (i.e. `1.0.0-beta1`) to trigger the `certification_upload` job.
This will allow the image to go through the Red Hat certification tests, but
will not release the images through the production channel (when that
functionality has been implemented).

It is also possible to add the `certification_upload` job to any pipeline
by setting the CI variable `PUSH_TO_REDHAT` to the value "true".
