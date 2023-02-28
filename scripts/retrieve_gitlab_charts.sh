#!/bin/bash
#
# This script executes during the image_build job of the pipeline
# and is responsible for retrieving the correct versions of the
# GitLab chart. These charts are then baked into the operator
# container image when the Dockerfile is processed.

set -eo pipefail

HELM="bin/helm"
GITLAB_CHART="gitlab/gitlab"
MAX_CHART_FETCH_ATTEMPTS=${MAX_CHART_FETCH_ATTEMPTS:-10}
CHART_FETCH_WAIT_TIME=${CHART_FETCH_WAIT_TIME:-30s}
DEV_API_CHARTS_PROJECT_URL="https://dev.gitlab.org/api/v4/projects/${DEV_CHARTS_PROJECT_ID}"

scripts_dir="$(dirname "$0")"
. "${scripts_dir}/add_gitlab_repo.sh"

build_from_source() {
    [ -z "${DEV_CHARTS_PAT}" -o -z "${DEV_CHARTS_PROJECT_ID}" ] && return 1

    local version="${1}"
    local source="/tmp/chart-${version}"

    rm -rf "${source}" && mkdir "${source}"
    curl -fSL \
        -H "PRIVATE-TOKEN:${DEV_CHARTS_PAT}" \
        "${DEV_API_CHARTS_PROJECT_URL}/repository/archive.tar.gz?sha=v${version}" | \
    tar -xzf - -C "${source}" --strip-component 1

    pushd "${source}"
    helm dependency update
    helm package --version="${version}" .
    popd

    mv "${source}/gitlab-${version}.tgz" ./charts && rm -rf "${source}"
}

# Download the charts to the charts directory
rm -rf charts && mkdir charts
for version in $(cat CHART_VERSIONS); do
    count=0
    echo "Fetching ${GITLAB_CHART}-${version}"
    while ! $($HELM fetch "${GITLAB_CHART}" --version "${version}" --destination ./charts/ 2> /dev/null); do
        if [ $count -ge "${MAX_CHART_FETCH_ATTEMPTS}" ]; then
            echo "  Fetch attempts exhausted. Attempting to build from source."
            if ! build_from_source "$version"; then
                echo "  Failed to build from source. Exiting."
                exit 1;
            else
              break
            fi
        fi

        echo "  Could not fetch chart. Sleeping for ${CHART_FETCH_WAIT_TIME} before attempting again."
        sleep "${CHART_FETCH_WAIT_TIME}"
        count=$((count+1))
    done
done
