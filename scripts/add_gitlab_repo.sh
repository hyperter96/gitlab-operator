#!/bin/sh -e

GITLAB_HELM_REPO="https://charts.gitlab.io/"
HELM="bin/helm"

scripts_dir="$(dirname "$0")"


add_gitlab_repo() {
    echo "Adding ${GITLAB_HELM_REPO} to list of helm repos"
    $HELM repo list | grep -q '^gitlab' || $HELM repo add gitlab ${GITLAB_HELM_REPO}
    $HELM repo update
}

. "${scripts_dir}/install_helm.sh"
add_gitlab_repo
