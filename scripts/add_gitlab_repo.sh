#!/bin/sh -e

GITLAB_HELM_REPO="https://charts.gitlab.io/"
GITLAB_CHART="gitlab/gitlab"
HELM=helm
HELM_VERSION="v3.4.1"

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
    HELM=$(find "$PWD"/ -name helm -type f -executable )
}

add_gitlab_repo() {
    echo "Adding ${GITLAB_HELM_REPO} to list of helm repos"
    $HELM repo list | grep -q '^gitlab' || $HELM repo add gitlab ${GITLAB_HELM_REPO}
    $HELM repo update
}


# Setup helm so that the GitLab charts can be fetched
if ! helm version --short > /dev/null 2>&1; then
  install_helm
fi
add_gitlab_repo
