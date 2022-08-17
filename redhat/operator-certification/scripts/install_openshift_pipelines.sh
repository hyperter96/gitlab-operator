#!/bin/sh

set -eu

# Based on https://docs.openshift.com/container-platform/4.9/cicd/pipelines/installing-pipelines.html

cleanup_files(){
    rm -rf ${MANIFEST}
}
trap cleanup_files EXIT

export RH_PIPELINES_CHANNEL="stable"

MANIFEST=$(mktemp)

create_manifest(){
    cat > ${MANIFEST} << EOF 
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: openshift-pipelines-operator
  namespace: openshift-operators
spec:
  channel: ${RH_PIPELINES_CHANNEL}
  name: openshift-pipelines-operator-rh 
  source: redhat-operators 
  sourceNamespace: openshift-marketplace 
EOF
}

apply_manifest(){
    kubectl apply -f ${MANIFEST}
}

save_manifest(){
    mv ${MANIFEST} openshift-pipelines.yaml
}

for cmd in "$@"
do
    $cmd
done