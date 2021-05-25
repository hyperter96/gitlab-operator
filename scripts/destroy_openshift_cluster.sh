#!/bin/bash -e

# Environment variables
CLUSTER_NAME="${CLUSTER_NAME:-ocp-$USER}"

GOOGLE_APPLICATION_CREDENTIALS="${GOOGLE_APPLICATION_CREDENTIALS:-gcloud.json}"
GOOGLE_CREDENTIALS="${GOOGLE_CREDENTIALS:-$(cat $GOOGLE_APPLICATION_CREDENTIALS)}"

LOG_LEVEL="${LOG_LEVEL:-info}"

INSTALL_DIR="install-${CLUSTER_NAME}"

export GOOGLE_CREDENTIALS  # needed for openshift-install to see them

echo 'Destroying cluster'
openshift-install destroy cluster \
  --dir "$INSTALL_DIR" --log-level "$LOG_LEVEL"
