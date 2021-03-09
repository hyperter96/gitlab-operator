#!/bin/bash -e

# Functional test that verifies the GitLab operator and CR install without error

NAMESPACE="${NAMESPACE:-gitlab-system}"
CLEANUP="${CLEANUP:-yes}"

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

  CHART_VERSION="${CHART_VERSION:-$(latest_chart_version)}"

  echo 'Starting test'
  install_required_operators
  install_crds
  install_gitlab_operator
  verify_operator_is_running
  install_gitlab_custom_resource
  verify_gitlab_is_running
}

install_required_operators() {
  # nginx-ingress
  echo 'Installing required operators'  # See https://www.itix.fr/blog/install-operator-openshift-cli/
  kubectl apply -f scripts/manifests/nginx-ingress-operator-group.yaml
  kubectl apply -f scripts/manifests/nginx-ingress-operator-sub.yaml
  wait_until_exists "crd/nginxingresscontrollers.k8s.nginx.org"
  kubectl wait --for=condition=Established crd/nginxingresscontrollers.k8s.nginx.org
  kubectl wait --for=condition=Available -n default deployment/nginx-ingress-operator

  # cert-manager
  kubectl apply -f scripts/manifests/cert-manager-sub.yaml
  wait_until_exists "crd/certmanagers.operator.cert-manager.io"
  kubectl wait --for=condition=Established crd/certmanagers.operator.cert-manager.io
  kubectl apply -f scripts/manifests/cert-manager-instance.yaml
  kubectl wait --for=condition=Initialized -n default certmanager/cert-manager
  kubectl wait --for=condition=Available -n default deployment/cert-manager-webhook
}

install_crds() {
  echo 'Installing operator CRDs'
  make install
}

install_gitlab_operator() {
  echo 'Installing GitLab operator'
  make suffix_clusterrolebinding_names
  make suffix_webhook_names
  make deploy
}

verify_operator_is_running() {
  echo 'Verifying that operator is running'

  kubectl wait --for=condition=Available -n "$NAMESPACE" deployment/gitlab-controller-manager

  local maxattempts=30
  local attempts=0
  until kubectl -n $NAMESPACE logs deployment/gitlab-controller-manager -c manager | grep -q 'successfully acquired lease'; do
    attempts=$((attempts+1))
    if [ "$attempts" -ge "$maxattempts" ]; then
      echo "Failed waiting for manager to start"; exit 1;
    fi
    echo -n '.'; sleep 2
  done
}

install_gitlab_custom_resource() {
  echo 'Installing GitLab custom resource'
  make CHART_VERSION="${CHART_VERSION}" deploy_sample
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
  kubectl -n $NAMESPACE exec deployment/gitlab-task-runner -- \
    bash -c 'curl -IL $GITLAB_WEBSERVICE_DEFAULT_PORT_8181_TCP_ADDR:8181'
}

cleanup() {
  echo 'Cleaning up test resources'
  kubectl delete namespace "$NAMESPACE"
  kubectl get clusterrolebindings -o=name | grep $NAMESPACE | xargs kubectl delete
  kubectl get validatingwebhookconfiguration -o name | grep $NAMESPACE | xargs kubectl delete
  kubectl get mutatingwebhookconfiguration -o name | grep $NAMESPACE | xargs kubectl delete
}

latest_chart_version() {
  local latest_chart_archive="$(basename $(ls charts/*.tgz | tail -n1) .tgz)"
  echo "${latest_chart_archive#gitlab-}"
}

wait_until_exists() {
  local resource="$1"
  local maxattempts="${2:-60}"
  local attempts=0
  local output
  local exitcode

  while true; do
    attempts=$((attempts+1))
    if [ "$attempts" -ge "$maxattempts" ]; then
      echo "Failed waiting for $resource"; exit 1;
    fi

    set +e
    output="$(kubectl -n "$NAMESPACE" get "$resource" 2>&1)"
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
