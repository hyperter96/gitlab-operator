#!/bin/sh

set -eu

# Based on https://docs.openshift.com/container-platform/4.9/cicd/pipelines/installing-pipelines.html

cleanup_files(){
    rm -rf ${MANIFEST}
}
trap cleanup_files EXIT

export RH_OCO_CHANNEL="alpha"
OCP_PROJECT=${OCP_PROJECT:-"gitlab-certification"}

MANIFEST=$(mktemp)

create_manifest(){
    cat > ${MANIFEST} << EOF 
---
apiVersion: operators.coreos.com/v1alpha2
kind: OperatorGroup
metadata:
  name: gitlab
  namespace: ${OCP_PROJECT}
spec:
  targetNamespaces:
  # - ${OCP_PROJECT}
---
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: operator-certification-operator
  namespace: ${OCP_PROJECT}
spec:
  channel: ${RH_OCO_CHANNEL}
  installPlanApproval: Automatic
  name: operator-certification-operator
  source: certified-operators
  sourceNamespace: openshift-marketplace
  # startingCSV: operator-certification-operator.v1.0.3
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