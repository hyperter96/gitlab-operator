#!/bin/sh -e
# Note: POSIX only! This script is expected to run in busybox ash.

# This script executes during the image_build job of the pipeline
# and is responsible for retrieving the correct versions of the
# GitLab chart. These charts are then baked into the operator
# container image when the Dockerfile is processed.

HELM="bin/helm"
GITLAB_CHART="gitlab/gitlab"

scripts_dir="$(dirname "$0")"
. "${scripts_dir}/add_gitlab_repo.sh"

# Download the charts to the charts directory
rm -rf charts && mkdir charts
for version in $(cat CHART_VERSIONS); do
    echo "Fetching ${GITLAB_CHART}-${version}"
    $HELM fetch "${GITLAB_CHART}" --version "${version}" --destination ./charts/
done
