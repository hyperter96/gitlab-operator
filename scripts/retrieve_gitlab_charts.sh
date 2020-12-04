#!/bin/sh -e

# This script executes during the image_build job of the pipeline
# and is responsible for retrieving the correct versions of the
# GitLab chart. These charts are then baked into the operator
# container image when the Dockerfile is processed.

GITLAB_HELM_REPO="https://charts.gitlab.io/"
GITLAB_CHART="gitlab/gitlab"
HELM_VERSION="v3.4.1"


chart_versions() {
    # escape the slash in the chart name
    # shellcheck disable=SC3060
    awk_chart_filter=${GITLAB_CHART//\//\\/}
    ./helm search repo ${GITLAB_CHART} -l 2>/dev/null | sed 1d | awk "
    /${awk_chart_filter}\s/ { print \$3 \":\" \$2 }"
}

install_helm() {
    echo "Installing helm to local directory"

    platform=""
    case $(uname -s) in
        Darwin)
            platform="darwin-amd64"
            ;;
        Linux)
            case $(uname -m) in
                x86_64)
                    platform="linux-amd64"
                    ;;
                aarch64)
                    platform="linux-arm64"
                    ;;
            esac
    esac
    HELM_RELEASE_URL="https://get.helm.sh/helm-${HELM_VERSION}-${platform}.tar.gz"
    wget -O - "${HELM_RELEASE_URL}" | tar xzf - ${platform}/helm
    mv ${platform}/helm .
    rm -rf ${platform}
}

add_gitlab_repo() {
    echo "Adding ${GITLAB_HELM_REPO} to list of helm repos"
    ./helm repo add gitlab ${GITLAB_HELM_REPO}
}

previous_minor() {
    # this will subtract 1 from the minor version.
    # if minor-1 is -1 return previous major-1 with now minor
    echo "$1" | awk -F'.' '{if (($2 - 1) == "-1")
                              print ($1 - 1) ".";
                            else
                              print $1 "." ($2 - 1)}'
}



# Setup helm so that the GitLab charts can be fetched
install_helm
add_gitlab_repo

# Retrieve a list of GitLab charts and determine versions to fetch
target_versions=""
next_version=""
for version_pair in $(chart_versions); do
    # Unpack the versions in to variables
    IFS=':' read -r gitlab_version chart_version <<EOF
$version_pair
EOF

    # Pick the first chart version if nothing selected yet
    if [ -z "$target_versions" ]; then
        target_versions=$chart_version
        next_version=$(previous_minor "$gitlab_version")
        continue
    fi

    if [ -n "$(expr "$gitlab_version" : "\($next_version\)")" ]; then
        target_versions="${target_versions}:$chart_version"
        next_version=$(previous_minor "$gitlab_version")
    fi

    # Only need to target 3 versions
    # shellcheck disable=SC2046
    if [ $(echo "${target_versions}" | awk -F: '{print NF - 1}') -eq 2 ]; then
        break
    fi
done

# download the target_versions charts to the charts directory
rm -rf charts && mkdir charts && cd charts
for version in $(echo "${target_versions}" | tr ':' ' '); do
    echo "Fetching ${GITLAB_CHART}-${version}"
    ../helm fetch "${GITLAB_CHART}" --version "${version}" 2>/dev/null
done
