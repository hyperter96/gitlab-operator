#!/bin/bash -e

# Functional test that verifies the GitLab operator and CR install without error

NAMESPACE="${NAMESPACE:-gitlab-system}"
CLEANUP="${CLEANUP:-yes}"
HOSTSUFFIX="${HOSTSUFFIX:-${NAMESPACE}}"
DOMAIN="${DOMAIN:-example.com}"
DEBUG="${DEBUG:-off}"

finish() {
  local exitcode=$?

  make restore_kustomize_files

  if [ $exitcode -ne 0 ]; then
    echo "!!!ERROR!!!"
    echo "deployment/gitlab-controller-manager logs"
    kubectl -n "$NAMESPACE" logs deployment/gitlab-controller-manager -c manager || true
  fi

  if [ "$CLEANUP" = "yes" ]; then
    cleanup
  else
    echo 'Skipping cleanup'
  fi
}
trap finish EXIT

main() {
  [ "$CLEANUP" = "only" ] && { cleanup; exit 0; }

  echo 'Starting test'
  install_crds
  install_gitlab_operator
  verify_operator_is_running
  install_gitlab_custom_resource
  copy_certificate
  verify_gitlab_is_running
}

# install_required_operators() {
#   echo 'Installing required operators'  # See https://www.itix.fr/blog/install-operator-openshift-cli/
#   make install_required_operators

#   # cert-manager
#   wait_until_exists "crd/certmanagers.operator.cert-manager.io"
#   kubectl wait --for=condition=Established crd/certmanagers.operator.cert-manager.io
#   kubectl wait --for=condition=Initialized -n default certmanager/cert-manager
#   kubectl wait --for=condition=Available -n default deployment/cert-manager-webhook
# }

install_crds() {
  echo 'Installing operator CRDs'
  make install_crds
}

install_gitlab_operator() {
  echo 'Installing GitLab operator'
  make suffix_clusterrolebinding_names
  make suffix_webhook_names
  make deploy_operator
}

verify_operator_is_running() {
  echo 'Verifying that operator is running'
  kubectl wait --for=condition=Available -n "$NAMESPACE" deployment/gitlab-controller-manager
}

install_gitlab_custom_resource() {
  echo 'Installing GitLab custom resource'
  make deploy_test_cr
}

copy_certificate() {
  echo 'Copying certificate to namespace'
  kubectl get secret -n default gitlab-ci-tls -o yaml \
    | yq eval 'del(.metadata.["namespace","resourceVersion","uid","annotations","creationTimestamp","selfLink","managedFields"])' - \
    | kubectl apply -n "$NAMESPACE" -f -
}

verify_gitlab_is_running() {
  echo 'Verifying that GitLab is running'

  statefulsets=(gitlab-gitaly gitlab-redis-master gitlab-minio gitlab-postgresql)
  wait_until_exists "statefulset/${statefulsets[0]}"
  for statefulset in "${statefulsets[@]}"; do
    kubectl -n "$NAMESPACE" rollout status -w --timeout 120s "statefulset/$statefulset"
    echo "statefulset/$statefulset ok"
  done

  deployments=(gitlab-gitlab-exporter gitlab-gitlab-shell gitlab-registry gitlab-sidekiq-all-in-1-v1 gitlab-task-runner gitlab-webservice-default gitlab-ingress-controller)
  wait_until_exists "deployment/${deployments[0]}"
  kubectl -n "$NAMESPACE" wait --timeout=600s --for condition=Available deployment -l app.kubernetes.io/managed-by=gitlab-operator

  echo 'Testing GitLab endpoint'
  curl -fIL "https://gitlab-$HOSTSUFFIX.$DOMAIN"
}

cleanup() {
  echo 'Cleaning up test resources'
  signal_failure=0

  # Turn off exit immediately if command fails so debug out can get generated
  set +e

  kubectl delete ns "$NAMESPACE"
  if [[ $? -ne 0 ]]; then
    signal_failure=1
  fi

  results=$(kubectl get clusterrolebindings -o=name | grep $NAMESPACE)
  [[ "$DEBUG" != "off" ]] && printf "** kubectl get clusterrolebinding results\n$results\n-----"
  echo "$results" | xargs kubectl delete

  results=$(kubectl get validatingwebhookconfiguration -o name | grep $NAMESPACE)
  [[ "$DEBUG" != "off" ]] && printf "** kubectl get validatingwebhookconfiguration results\n$results\n-----"
  echo "$results" | xargs kubectl delete
  if [[ $? -ne 0 ]]; then
    signal_failure=1
  fi

  if [[ $signal_failure -eq 1 ]]; then
    exit 1
  fi

  # Turn back on to exit immediately
  set -e
}

wait_until_exists() {
  local resource="$1"
  local namespace="${2:-$NAMESPACE}"
  local maxattempts="${3:-60}"
  local attempts=0
  local output
  local exitcode

  while true; do
    attempts=$((attempts+1))
    if [ "$attempts" -ge "$maxattempts" ]; then
      echo "Failed waiting for $resource"; exit 1;
    fi

    set +e
    output="$(kubectl -n "$namespace" get "$resource" 2>&1)"
    exitcode=$?
    set -e
    if [ $exitcode -eq 0 ]; then
      break
    fi

    if [[ "$output" == *"not found"* ]]; then
      echo -n '.'; sleep 2
    else
      echo "$output"; exit 1
    fi
  done
}

main
