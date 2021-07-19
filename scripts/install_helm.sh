#!/bin/bash -e

HELM_VERSION="v3.6.3"
BINDIR='bin'


install_helm() {
    echo "Installing helm to local directory $BINDIR/"

    local platform=""
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

    mkdir -p $BINDIR
    mv "${platform}/helm" "${BINDIR}/helm"
    rm -rf "${platform}"
}

if ! $BINDIR/helm version --short > /dev/null 2>&1; then
  install_helm
fi
