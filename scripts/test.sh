#!/bin/bash -e

# Smoke test that verifies the GitLab operator installs without error

NAMESPACE="${NAMESPACE:-gitlab-system}"
CLEANUP="${CLEANUP:-yes}"

finish() {
  git checkout -q config/manager  # Restore manager image tag
  git checkout -q config/default/kustomization.yaml  # Restore namespace
  git checkout -q config/rbac  # Restore ClusterRoleBinding names

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
  install_required_operators
  install_crds
  rename_clusterrolebindings
  install_gitlab_operator
  verify_operator_running
}

install_required_operators() {
  # nginx-ingress
  echo 'Installing required operators'  # See https://www.itix.fr/blog/install-operator-openshift-cli/
  kubectl apply -f scripts/manifests/nginx-ingress-operator-group.yaml
  kubectl apply -f scripts/manifests/nginx-ingress-operator-sub.yaml
  until kubectl get crd nginxingresscontrollers.k8s.nginx.org &>/dev/null; do
    echo -n '.'
    sleep 1
  done
  kubectl wait --for=condition=Established crd/nginxingresscontrollers.k8s.nginx.org
  kubectl wait --for=condition=Available -n default deployment/nginx-ingress-operator

  # cert-manager
  kubectl apply -f scripts/manifests/cert-manager-sub.yaml
  until kubectl get crd certmanagers.operator.cert-manager.io &>/dev/null; do
    echo -n '.'
    sleep 1
  done
  kubectl wait --for=condition=Established crd/certmanagers.operator.cert-manager.io
  kubectl apply -f scripts/manifests/cert-manager-instance.yaml
  kubectl wait --for=condition=Initialized -n default certmanager/cert-manager
  kubectl wait --for=condition=Available -n default deployment/cert-manager-webhook
}

install_crds() {
  echo 'Installing operator CRDs'
  make install
}

rename_clusterrolebindings() {
  echo 'Renaming ClusterRoleBindings'

  local manifests=($(grep -rwl 'config/rbac' -e 'kind: ClusterRoleBinding'))

  for m in "${manifests[@]}"; do
    local currentName="$(yq read "$m" 'metadata.name')"
    yq write --inplace "${m}" 'metadata.name' "$NAMESPACE-$currentName"
  done
}

install_gitlab_operator() {
  echo 'Installing GitLab operator'
  make deploy
}

verify_operator_running() {
  echo 'Verifying that operator is running'

  kubectl wait --for=condition=Available -n "$NAMESPACE" deployment/gitlab-controller-manager
}

cleanup() {
  echo 'Cleaning up test resources'
  local clusterrolebindings=($(kubectl get clusterrolebindings -o=name | grep $NAMESPACE))
  for crb in "${clusterrolebindings[@]}"; do
    kubectl delete "${crb}"
  done

  kubectl delete namespace "$NAMESPACE"
}

main
