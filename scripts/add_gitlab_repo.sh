#!/bin/sh -e

HELM="bin/helm"
scripts_dir="$(dirname "$0")"
GITLAB_HELM_REPO="https://charts.gitlab.io/"
HELM_ARGS=""

is_dev_mirror() {
    [ "$CI_PROJECT_PATH" = "gitlab/cloud-native/gitlab-operator" ]
}

add_gitlab_repo() {
    echo "Adding ${GITLAB_HELM_REPO} to list of helm repos"
    $HELM repo list | grep -q '^gitlab' || $HELM repo add ${HELM_ARGS} gitlab ${GITLAB_HELM_REPO}
    $HELM repo update
}

. "${scripts_dir}/install_helm.sh"

if is_dev_mirror; then
    CHANNEL=${CHANNEL:-stable}
    GITLAB_HELM_REPO="${CI_API_V4_URL}/projects/${DEV_CHARTS_PROJECT_ID}/packages/helm/${CHANNEL}"
    HELM_ARGS="--username ${DEV_CHARTS_USERNAME} --password ${DEV_CHARTS_PAT}"
fi

add_gitlab_repo
