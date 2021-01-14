#!/bin/bash -e

# Environment variables
CLUSTER_NAME="${CLUSTER_NAME:-ocp-$USER}"
BASE_DOMAIN="${BASE_DOMAIN:-k8s-ft.win}"
GCP_PROJECT_ID="${GCP_PROJECT_ID:-cloud-native-182609}"
GCP_REGION="${GCP_REGION:-us-central1}"

GOOGLE_APPLICATION_CREDENTIALS="${GOOGLE_APPLICATION_CREDENTIALS:-gcloud.json}"
GOOGLE_CREDENTIALS="${GOOGLE_CREDENTIALS:-$(cat $GOOGLE_APPLICATION_CREDENTIALS)}"

PULL_SECRET_FILE="${PULL_SECRET_FILE:-pull_secret}"
PULL_SECRET="${PULL_SECRET:-$(cat $PULL_SECRET_FILE)}"

SSH_PUBLIC_KEY_FILE="${SSH_PUBLIC_KEY_FILE:-$HOME/.ssh/id_rsa.pub}"
SSH_PUBLIC_KEY="${SSH_PUBLIC_KEY:-$(cat $SSH_PUBLIC_KEY_FILE)}"

LOG_LEVEL="${LOG_LEVEL:-info}"

INSTALL_DIR="install-${CLUSTER_NAME}"

main() {
  export GOOGLE_CREDENTIALS  # needed for openshift-install to see them

  verify_requirements
  render_config_file
  create_cluster
}

verify_requirements() {
  echo 'Verifying requirements'

  _verify_installed 'openshift-install'
  _verify_installed 'oc'
}

render_config_file() {
  echo 'Rendering install-config file'
  template_data="$(cat scripts/install-config.template.yaml)"

  template_data="$(echo "${template_data//CLUSTER_NAME/$CLUSTER_NAME}")"
  template_data="$(echo "${template_data//PULL_SECRET/$PULL_SECRET}")"
  template_data="$(echo "${template_data//SSH_PUBLIC_KEY/$SSH_PUBLIC_KEY}")"
  template_data="$(echo "${template_data//BASE_DOMAIN/$BASE_DOMAIN}")"
  template_data="$(echo "${template_data//GCP_PROJECT_ID/$GCP_PROJECT_ID}")"
  template_data="$(echo "${template_data//GCP_REGION/$GCP_REGION}")"

  mkdir -p $INSTALL_DIR
  echo "$template_data" > "$INSTALL_DIR/install-config.yaml"
}

create_cluster() {
  echo "Creating cluster '$CLUSTER_NAME'"
  openshift-install create cluster \
    --dir "$INSTALL_DIR" --log-level "$LOG_LEVEL"
}

_verify_installed() {
  local cmd="$1"
  if ! type "$cmd" > /dev/null; then
    echo "'$cmd' tool is required and not installed"
    exit 1
  fi
}

main
