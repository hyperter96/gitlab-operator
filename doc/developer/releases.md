# Versioning

The GitLab Operator uses [semver versioning](https://semver.org/).
Version tags should [be the semver version string](../adr/0009-version-tagging.md).

## Documentation

Operator documentation is available in the `doc/` directory.

## Red Hat Certification

The release pipeline will contain a `certification_upload` job when the
repository has been tagged with a semver version (i.e. `1.0.0`). This job
will trigger the Red Hat API to request the image be passed through
Red Hat's certification pipeline. The results of the certification pipeline
are published through Red Hat's Connect portal.

It is also possible to pass a release candidate tag (i.e. `1.0.0-rc1`) or a
beta tag (i.e. `1.0.0-beta1`) to trigger the `certification_upload` job.
This will allow the image to go through the Red Hat certification tests, but
will not release the images through the production channel (when that
functionality has been implemented).

It is also possible to add the `certification_upload` job to any pipeline
by setting the CI variable `REDHAT_CERTIFICATION` to the value "true".

In addition, it is possible to run the `scripts/redhat_certification.rb`
script and query the Red Hat API for the status of scan requests that have
been submitted. Executing `scripts/redhat_certification.rb -s` will display
a list of images and their current status in the Red Hat certification
pipeline.

In order to execute the script independently from GitLab CI one needs to
create the `REDHAT_API_TOKEN` environmental variable. This variable is set
to the personal token generated on the [Connect portal](https://connect.redhat.com/account/api-keys).
The token used by GitLab CI is stored in the 1Password Build vault under the
"Red HatCertification Token" entry.
