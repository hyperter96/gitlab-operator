# Versioning

The GitLab OpenShift operator uses [semver versioning](https://semver.org/).
Version tags should be prefixed with the letter `v` followed by the semver
version string.

# Red Hat Certification

The release pipeline will contain a `certification_upload` job when the
repository has been tagged with a semver version prefixed with a 'v'
(i.e. `v1.0.0`). This job will upload the operator image to Red Hat for
certification test and allow the operator to be published through the
Red Hat Connect portal.

It is also possible to pass a release candidate tag (i.e. `v1.0.0-rc1`) or  a
beta tag (i.e. `v1.0.0-beta1`) to trigger the `certification_upload` job.
This will allow the image to go through the Red Hat certification tests, but
will not release the images through the production channel (when that
functionality has been implemented).

It is also possible to add the `certification_upload` job to any pipeline
by setting the CI variable `PUSH_TO_REDHAT` to the value "true".
