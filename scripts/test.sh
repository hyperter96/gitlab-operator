#!/bin/bash -e

# Functional test that verifies the GitLab operator and CR install without error

TESTS_NAMESPACE="${TESTS_NAMESPACE:-gitlab-system}"
CLEANUP="${CLEANUP:-yes}"
HOSTSUFFIX="${HOSTSUFFIX:-${TESTS_NAMESPACE}}"
DOMAIN="${DOMAIN:-example.com}"
DEBUG_CLEANUP="${DEBUG_CLEANUP:-off}"
KUBECTL_WAIT_TIMEOUT_SECONDS=${KUBECTL_WAIT_TIMEOUT_SECONDS:-"600s"}

REGISTRY_AUTH_SECRET_NS=${REGISTRY_AUTH_SECRET_NS:-""}
REGISTRY_AUTH_SECRET=${REGISTRY_AUTH_SECRET:-""}

# When defined - skip cleanup at the end of script run
NO_TRAP=${NO_TRAP:-""}

# Command for `yq`, expected to be https://github.com/mikefarah/yq
YQ=${YQ:-"yq"}

export IMG TAG NAMESPACE=${TESTS_NAMESPACE}
PLATFORM="${PLATFORM:-kubernetes}"

finish() {
  local exitcode=$?

  make restore_kustomize_files

  if [ $exitcode -ne 0 ]; then
    echo "!!!ERROR!!!"
    echo "deployment/gitlab-controller-manager logs"
    kubectl -n "$TESTS_NAMESPACE" logs deployment/gitlab-controller-manager -c manager || true
  fi

  if [ "$CLEANUP" = "yes" ]; then
    cleanup
  else
    echo 'Skipping cleanup'
  fi
}
[ -z "${NO_TRAP}" ] && trap finish EXIT

main() {
  [ "$CLEANUP" = "only" ] && { cleanup; exit 0; }

  echo 'Starting test'
  if [ "${CI}" == "true" ] 
  then
    # called from pipeline
    # all artifacts should be in place

    build_kustomized

    deploy_kustomized
  else
    install_crds
    create_namespace
    install_gitlab_operator
    verify_operator_is_running
    copy_certificate
    build_gitlab_custom_resource
    install_gitlab_custom_resource
    verify_gitlab_is_running
  fi
}

build_kustomized(){
  create_kustomization
  setup_kustomization
  compile_kustomization
  build_gitlab_custom_resource
}

deploy_kustomized(){
  create_namespace
  setup_kustomization
  install_kustomization

  verify_operator_is_running
  copy_certificate
  install_gitlab_custom_resource
  verify_gitlab_is_running
}

_repurpose_cr(){
  # Strip all the metadata k8s adds to resource
  # upon creation and make resource more "generic"
  ${YQ} eval "del(.metadata.namespace,.metadata.creationTimestamp,.metadata.resourceVersion,.metadata.selfLink,.metadata.uid,.metadata.managedFields)" $@
}

create_namespace() {
  kubectl get namespace ${TESTS_NAMESPACE} > /dev/null 2>&1 || kubectl create namespace ${TESTS_NAMESPACE}
  if [ -n "${REGISTRY_AUTH_SECRET}" ] && [ -n "${REGISTRY_AUTH_SECRET_NS}" ]
  then
    kubectl get secret ${REGISTRY_AUTH_SECRET} --namespace=${REGISTRY_AUTH_SECRET_NS} -o yaml \
      | _repurpose_cr - \
      | sed -e "s/namespace: ${REGISTRY_AUTH_SECRET_NS}/namespace: ${TESTS_NAMESPACE}/" \
      | kubectl apply --namespace=${TESTS_NAMESPACE} -f -
  fi
}

install_crds() {
  #TODO Deprecate install_crds
  echo 'Installing operator CRDs'
  make install_crds
}


create_kustomization() {
  # create kustomization infrastructure
  for d in generic openshift
  do
    mkdir -p .build/kustomize/${d}
    cp -r config/ci/* .build/kustomize/${d}/
    cp .build/operator.yaml .build/kustomize/${d}
  done
  cp .build/openshift_resources.yaml .build/kustomize/openshift/

  (
    cd .build/kustomize/generic
    if [ -n "${REGISTRY_AUTH_SECRET}" ]
    then
      ${YQ} eval -i ".spec.template.spec.imagePullSecrets[0].name=\"${REGISTRY_AUTH_SECRET}\"" patches/dev-pullSecret.yaml
      kustomize edit add patch --kind Deployment --path patches/dev-pullSecret.yaml
    fi
    kustomize edit set image ${IMG}:${TAG}
    kustomize edit set namesuffix -- "-${TESTS_NAMESPACE}"
    kustomize edit set namespace "${TESTS_NAMESPACE}"
  )
  (
    cd .build/kustomize/openshift
    if [ -n "${REGISTRY_AUTH_SECRET}" ]
    then
      ${YQ} eval -i ".spec.template.spec.imagePullSecrets[0].name=\"${REGISTRY_AUTH_SECRET}\"" patches/dev-pullSecret.yaml
      kustomize edit add patch --kind Deployment --path patches/dev-pullSecret.yaml
    fi
    kustomize edit set image ${IMG}:${TAG}
    kustomize edit set namesuffix -- "-${TESTS_NAMESPACE}"
    kustomize edit set namespace "${TESTS_NAMESPACE}"
    kustomize edit add resource openshift_resources.yaml
  )
}

setup_kustomization() {
  local basedir
  echo "Setting up environment for kustomize"
  if [ -n "$1" ]
  then
    basedir=$1
  else
    basedir=$(pwd)
  fi
  DEPLOYMENT_DIR=${basedir}/.install
  BUILD_DIR=${basedir}/.build

  if [ "${PLATFORM}" == "openshift" ]
  then
    MANIFEST_DIR=${basedir}/.build/kustomize/openshift
  else
    MANIFEST_DIR=${basedir}/.build/kustomize/generic
  fi
  mkdir -p ${DEPLOYMENT_DIR}
  mkdir -p ${BUILD_DIR}
}

compile_kustomization() {
  # Needs to have setup_kustomization to be ran first
  echo "Compiling kustomize'd manifest"
  pushd ${MANIFEST_DIR}
  kustomize build > deployment.yaml
  set -x
  cp deployment.yaml ${BUILD_DIR}/glop-${HOSTSUFFIX}.${DOMAIN}.yaml
  set +x
  popd
}

install_kustomization() {
  echo "Deploying operator"
  set -x
  kubectl apply -f ${BUILD_DIR}/glop-${HOSTSUFFIX}.${DOMAIN}.yaml
  cp ${BUILD_DIR}/glop-${HOSTSUFFIX}.${DOMAIN}.yaml ${DEPLOYMENT_DIR}/glop-${HOSTSUFFIX}.${DOMAIN}.yaml
  set +x
}

install_gitlab_operator() {
  echo 'Installing GitLab operator'
  make suffix_clusterrolebinding_names
  make suffix_webhook_names
  if [ "$PLATFORM" == "openshift" ]; then
    make deploy_openshift_resources
  fi
  make deploy_operator
}

verify_operator_is_running() {
  echo 'Verifying that operator is running'
  kubectl wait --for=condition=Available -n "$TESTS_NAMESPACE" deployment/gitlab-controller-manager
}

build_gitlab_custom_resource() {
  echo 'Building GitLab custom resource manifest'
  make build_test_cr
  set -x
  YQ_CMD="."
  [ -n "${REGISTRY_AUTH_SECRET}" ] && \
    kubectl get secret --namespace="${TESTS_NAMESPACE}" "${REGISTRY_AUTH_SECRET}" && \
    YQ_CMD=".spec.chart.values.global.image.pullSecrets[0].name=\"${REGISTRY_AUTH_SECRET}\""
  ${YQ} eval "${YQ_CMD}" .build/test_cr.yaml > ${BUILD_DIR}/gitlab-${HOSTSUFFIX}.${DOMAIN}.yaml
  set +x
}

install_gitlab_custom_resource() {
  # requres "build_gitlab_custom_resource" to be ran first
  echo 'Installing GitLab custom resource'
  set -x
  kubectl apply -n ${TESTS_NAMESPACE} -f ${BUILD_DIR}/gitlab-${HOSTSUFFIX}.${DOMAIN}.yaml
  cp ${BUILD_DIR}/gitlab-${HOSTSUFFIX}.${DOMAIN}.yaml ${DEPLOYMENT_DIR}/gitlab-${HOSTSUFFIX}.${DOMAIN}.yaml
  set +x
}

copy_certificate() {
  echo 'Copying certificate to namespace'
  kubectl get secret -n default gitlab-ci-tls -o yaml \
    | ${YQ} eval 'del(.metadata.["namespace","resourceVersion","uid","annotations","creationTimestamp","selfLink","managedFields"])' - \
    | kubectl apply -n "$TESTS_NAMESPACE" -f -
}

verify_gitlab_is_running() {
  echo 'Verifying that GitLab is running'

  statefulsets=(gitlab-gitaly gitlab-redis-master gitlab-minio gitlab-postgresql)
  wait_until_exists "statefulset/${statefulsets[0]}"
  for statefulset in "${statefulsets[@]}"; do
    kubectl -n "$TESTS_NAMESPACE" rollout status -w --timeout 120s "statefulset/$statefulset"
    echo "statefulset/$statefulset ok"
  done

  echo 'Waiting for Migrations...'
  sleep 5
  kubectl -n "$TESTS_NAMESPACE" wait --timeout="${KUBECTL_WAIT_TIMEOUT_SECONDS}" --for condition=Complete job -l app=migrations

  echo 'Waiting for Deployments...'
  sleep 5
  kubectl -n "$TESTS_NAMESPACE" wait --timeout="${KUBECTL_WAIT_TIMEOUT_SECONDS}" --for condition=Available deployment -l app.kubernetes.io/managed-by=gitlab-operator

  local endpoint="https://gitlab-$HOSTSUFFIX.$DOMAIN"
  echo "Testing GitLab endpoint: $endpoint"
  sleep 5
  curl --retry 5 --retry-delay 10 -fIL "$endpoint"
}

cleanup() {
  echo 'Cleaning up test resources'
  signal_failure=0

  # Turn off exit immediately if command fails so debug out can get generated
  set +e

  if [ "${CI}" == "true" ]
  then
    # make sure we know where manifest is:
    setup_kustomization

    set -x
    # delete CR
    kubectl delete -f ${BUILD_DIR}/gitlab-${HOSTSUFFIX}.${DOMAIN}.yaml
    # delete operator resources
    kubectl delete -f ${BUILD_DIR}/glop-${HOSTSUFFIX}.${DOMAIN}.yaml
    kubectl delete ns "$TESTS_NAMESPACE"
    set +x
  else
    set -x
    kubectl delete ns "$TESTS_NAMESPACE"
    set +x
    if [[ $? -ne 0 ]]; then
      signal_failure=1
    fi

    results=$(kubectl get clusterrolebindings -o=name | grep $TESTS_NAMESPACE)
    [[ "$DEBUG_CLEANUP" != "off" ]] && printf "** kubectl get clusterrolebinding results\n$results\n-----"
    echo "$results" | xargs kubectl delete

    results=$(kubectl get validatingwebhookconfiguration -o name | grep $TESTS_NAMESPACE)
    [[ "$DEBUG_CLEANUP" != "off" ]] && printf "** kubectl get validatingwebhookconfiguration results\n$results\n-----"
    echo "$results" | xargs kubectl delete
  fi

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
  local namespace="${2:-$TESTS_NAMESPACE}"
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

# main
if [ "$#" -lt 1 ]
then
  main
else
  for cmd in "$@"
  do
    $cmd
  done
fi