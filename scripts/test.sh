#!/bin/bash -e

# Functional test that verifies the GitLab operator and CR install without error

TESTS_NAMESPACE="${TESTS_NAMESPACE:-gitlab-system}"
CLEANUP="${CLEANUP:-yes}"
HOSTSUFFIX="${HOSTSUFFIX:-${TESTS_NAMESPACE}}"
DOMAIN="${DOMAIN:-example.com}"
DEBUG_CLEANUP="${DEBUG_CLEANUP:-off}"

REGISTRY_AUTH_SECRET_NS=${REGISTRY_AUTH_SECRET_NS:-""}
REGISTRY_AUTH_SECRET=${REGISTRY_AUTH_SECRET:-""}

BASE_DIR=${BASE_DIR:-$(pwd)}
export INSTALL_DIR=$(realpath ${INSTALL_DIR:-"${BASE_DIR}/.install"})
export BUILD_DIR=$(realpath ${BUILD_DIR:-"${BASE_DIR}/.build"})

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
    mkdir -p ${BUILD_DIR}/kustomize/${d}
    cp -r config/ci/* ${BUILD_DIR}/kustomize/${d}/
    cp ${BUILD_DIR}/operator.yaml ${BUILD_DIR}/kustomize/${d}
  done
  cp ${BUILD_DIR}/openshift_resources.yaml ${BUILD_DIR}/kustomize/openshift/

  (
    cd ${BUILD_DIR}/kustomize/generic
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
    cd ${BUILD_DIR}/kustomize/openshift
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
  if [ "${PLATFORM}" == "openshift" ]
  then
    MANIFEST_DIR=${BUILD_DIR}/kustomize/openshift
  else
    MANIFEST_DIR=${BUILD_DIR}/kustomize/generic
  fi
  mkdir -p ${INSTALL_DIR}
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
  cp ${BUILD_DIR}/glop-${HOSTSUFFIX}.${DOMAIN}.yaml ${INSTALL_DIR}/glop-${HOSTSUFFIX}.${DOMAIN}.yaml
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
  ${YQ} eval "${YQ_CMD}" ${BUILD_DIR}/test_cr.yaml  > ${BUILD_DIR}/gitlab-${HOSTSUFFIX}.${DOMAIN}.yaml
  [ ${TESTS_NAMESPACE} != "gitlab-system" ] \
    && ${YQ} -i eval ".spec.chart.values.global.ingress.class=\"gitlab-nginx-${TESTS_NAMESPACE}\"" \
          ${BUILD_DIR}/gitlab-${HOSTSUFFIX}.${DOMAIN}.yaml
  set +x
}

install_gitlab_custom_resource() {
  # requres "build_gitlab_custom_resource" to be ran first
  echo 'Installing GitLab custom resource'
  set -x
  kubectl apply -n ${TESTS_NAMESPACE} -f ${BUILD_DIR}/gitlab-${HOSTSUFFIX}.${DOMAIN}.yaml
  cp ${BUILD_DIR}/gitlab-${HOSTSUFFIX}.${DOMAIN}.yaml ${INSTALL_DIR}/gitlab-${HOSTSUFFIX}.${DOMAIN}.yaml
  set +x
}

copy_certificate() {
  echo 'Copying certificate to namespace'
  kubectl get secret -n default gitlab-ci-tls -o yaml \
    | ${YQ} eval 'del(.metadata.["namespace","resourceVersion","uid","annotations","creationTimestamp","selfLink","managedFields"])' - \
    | kubectl apply -n "$TESTS_NAMESPACE" -f -
}

verify_gitlab_is_running() {
  wait_until_gitlab_running
  test_gitlab_endpoint
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
    # delete operator resources (except CustomResourceDefinition)
    CRD_INDEX=$(${YQ} eval 'select(.kind== "CustomResourceDefinition") | documentIndex' ${BUILD_DIR}/glop-${HOSTSUFFIX}.${DOMAIN}.yaml)
    ${YQ} eval "select(documentIndex != $CRD_INDEX)" ${BUILD_DIR}/glop-${HOSTSUFFIX}.${DOMAIN}.yaml > ${BUILD_DIR}/glop-${HOSTSUFFIX}.${DOMAIN}.no-crd.yaml
    kubectl delete -f ${BUILD_DIR}/glop-${HOSTSUFFIX}.${DOMAIN}.no-crd.yaml
    # delete namespace
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

wait_until_gitlab_running() {
  local sleepSeconds=10
  local maxattempts=120
  local attempts=0
  local exitcode
  local output

  echo 'Verifying that GitLab is running'

  while true; do
    output="$(kubectl -n "$TESTS_NAMESPACE" get gitlab/gitlab -ojsonpath='{.status.phase}' 2>&1)"
    exitcode=$?

    if [ $exitcode -ne 0 ]; then
      echo "$output"; exit $exitcode
    fi

    attempts=$((attempts+1))
    if [ "$attempts" -ge "$maxattempts" ]; then
      echo "Failed waiting for GitLab to be Running, current status is $output"; exit 1;
    fi

    if [[ "$output" == 'Running' ]]; then
      break
    else
      echo -n '.'; sleep $sleepSeconds
    fi
  done
}

test_gitlab_endpoint() {
  local endpoint="https://gitlab-$HOSTSUFFIX.$DOMAIN"

  echo "Testing GitLab endpoint: $endpoint"
  sleep 5
  curl --retry 5 --retry-delay 10 -fIL "$endpoint"
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
