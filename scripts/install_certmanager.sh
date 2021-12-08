#!/bin/bash -e

PLATFORM="${1:-openshift}"

CLUSTER_NAME="${CLUSTER_NAME:-ocp-$USER}"
INSTALL_DIR="install-${CLUSTER_NAME}"
KUBECONFIG="${KUBECONFIG:-$INSTALL_DIR/auth/kubeconfig}"
BASE_DOMAIN="${BASE_DOMAIN:-k8s-ft.win}"
GCP_PROJECT_ID="${GCP_PROJECT_ID:-cloud-native-182609}"
GOOGLE_APPLICATION_CREDENTIALS="${GOOGLE_APPLICATION_CREDENTIALS:-gcloud.json}"
GOOGLE_CREDENTIALS="${GOOGLE_CREDENTIALS:-$(cat $GOOGLE_APPLICATION_CREDENTIALS)}"

HELM="bin/helm"

install_certmanager() {
  echo 'Installing cert-manager'

  local scripts_dir="$(dirname "$0")"

  . "${scripts_dir}/install_helm.sh"

  $HELM repo list | grep -q '^jetstack' || $HELM repo add jetstack https://charts.jetstack.io
  $HELM repo update

  $HELM upgrade --install \
    cert-manager-helm jetstack/cert-manager \
    --namespace default \
    --version v1.1.0 \
    --values "scripts/manifests/cert-manager-values-$PLATFORM.yaml"

  sleep 10

  export KUBECONFIG  # needed so kubectl apply uses the correct cluster

  local google_credentials_json="$(echo $GOOGLE_CREDENTIALS | base64)"

  template_data="$(cat scripts/manifests/cert-manager-$PLATFORM.yaml)"
  template_data="$(echo "${template_data//GOOGLE_CREDENTIALS/$google_credentials_json}")"
  template_data="$(echo "${template_data//GCP_PROJECT_ID/$GCP_PROJECT_ID}")"
  template_data="$(echo "${template_data//CLUSTER_NAME/$CLUSTER_NAME}")"
  template_data="$(echo "${template_data//BASE_DOMAIN/$BASE_DOMAIN}")"

  echo "$template_data" | kubectl apply -f -
}

install_certmanager
