#!/bin/bash

set -eu


OC=${OC:-oc}
TKN=${TKN:-tkn}

BUILD=${BUILD:-".build"}

OCP_PROJECT=${OCP_PROJECT:-"gitlab-certification"}

SUBMIT=${SUBMIT:-"true"}

###
## Git-related variables:
GIT_BRANCH=${GIT_BRANCH:-"main"}
## Note: Use SSH-based Git URI
# GIT_FORK_REPO_URL="git@github.com:XXXX/certified-operators.git"
# GIT_USERNAME=
# GIT_EMAIL=
# GITHUB_TOKEN
# SSH_KEY_FILE

###
## RedHat-specific:
# PYXIS_API_KEY_FILE

# OPERATOR_BUNDLE_PATH=
# KUBECONFIG=

export KUBECONFIG
SCRIPT_DIR=$(dirname $(realpath $0))
PIPELINES_REPO="https://github.com/redhat-openshift-ecosystem/operator-pipelines"
UPSTREAM_REPO_NAME="redhat-openshift-ecosystem/certified-operators"

BUILD=$(realpath ${BUILD})

help(){
    echo "Required Environment Variables with sample values:"
    set +u
    for env_var_spec in \
            GIT_FORK_REPO_URL:git@github.com:XXXX/certified-operators.git \
            GIT_USERNAME:foo \
            GIT_EMAIL:foo@gitlab.com \
            GITHUB_TOKEN_FILE:/path/to/github_token.txt \
            SSH_KEY_FILE:/path/to/ssh/key \
            PYXIS_API_KEY_FILE:/path/to/pyxis/key \
            OPERATOR_BUNDLE_PATH:operators/gitlab-operator-kubernetes/0.9.1 \
            KUBECONFIG:/path/to/kubeconfig
    do
        env_var=${env_var_spec%%:*}
        env_var_value=${!env_var}
        env_var_text=${env_var_spec#*:}
        echo -e "  $env_var =\t\"$env_var_value\" # ($env_var_text)"
    done
    set -u
    echo "Commands:"
    # double-escaping grep search to avoid self-detection
    for c in $(grep -F '(''){' redhat/operator-certification/scripts/operator_certification_pipeline.sh | sed -e 's#()''{##g' | grep -ve '^_')
    do
        echo "  $c"
    done
}

_must_have(){
    # Make sure that variables that do not have default values are set
    set +u
    for varname in "$@"
    do
        if [ -z "${!varname}" ]; then
            echo "$varname is undefined"
            echo ""
            help
            exit 1
        fi
    done
    set -u
}

create_pipeline_project(){
    _must_have KUBECONFIG OC OCP_PROJECT
    $OC adm new-project ${OCP_PROJECT}
    set_project
}

set_project(){
    _must_have KUBECONFIG OC OCP_PROJECT
    $OC project ${OCP_PROJECT}
}

create_kubeconfig_secret(){
    _must_have KUBECONFIG OC
    $OC create secret generic kubeconfig --from-file=kubeconfig=${KUBECONFIG}
}

import_catalogs(){
    _must_have KUBECONFIG OC
    $OC import-image certified-operator-index \
    --from=registry.redhat.io/redhat/certified-operator-index \
    --reference-policy local \
    --scheduled \
    --confirm \
    --all

    $OC import-image redhat-marketplace-index \
    --from=registry.redhat.io/redhat/redhat-marketplace-index \
    --reference-policy local \
    --scheduled \
    --confirm \
    --all
}

fetch_certification_pipeline(){
    pushd ${BUILD}
    if [ ! -d "operator-pipelines" ]; then 
        git clone ${PIPELINES_REPO}
    else
        (cd operator-pipelines; git pull)
    fi
    popd
}

install_certification_pipeline(){
    _must_have KUBECONFIG
    pushd ${BUILD}/operator-pipelines
    $OC apply -R -f ansible/roles/operator-pipeline/templates/openshift/pipelines
    $OC apply -R -f ansible/roles/operator-pipeline/templates/openshift/tasks
    popd
}

create_secrets(){
    _must_have GITHUB_TOKEN_FILE PYXIS_API_KEY_FILE SSH_KEY_FILE
    $OC create secret generic github-api-token --from-file GITHUB_TOKEN=${GITHUB_TOKEN_FILE}
    $OC create secret generic pyxis-api-secret --from-file pyxis_api_key=${PYXIS_API_KEY_FILE}
    # Needed during digest pinning
    $OC create secret generic github-ssh-credentials --from-file id_rsa=${SSH_KEY_FILE}
}

run_certification_pipeline_manual(){
    _must_have TKN GIT_FORK_REPO_URL OPERATOR_BUNDLE_PATH GIT_USERNAME GIT_EMAIL
    pushd ${BUILD}/operator-pipelines
    # this worked without pin_digests... but resulted in failing pipeline upstream
    $TKN pipeline start operator-ci-pipeline \
        --use-param-defaults \
        --param git_repo_url=${GIT_FORK_REPO_URL} \
        --param git_branch=${GIT_BRANCH} \
        --param upstream_repo_name=${UPSTREAM_REPO_NAME} \
        --param bundle_path=${OPERATOR_BUNDLE_PATH} \
        --param env=prod \
        --workspace name=pipeline,volumeClaimTemplateFile=templates/workspace-template.yml \
        --workspace name=ssh-dir,secret=github-ssh-credentials \
        --showlog \
        --param git_username=${GIT_USERNAME} \
        --param git_email=${GIT_EMAIL} \
        --param tlsverify=false \
        --param pin_digests=true \
        --param submit=true
    popd
}

run_certification_pipeline_automated(){
    _must_have KUBECONFIG TKN GIT_FORK_REPO_URL OPERATOR_BUNDLE_PATH GIT_USERNAME GIT_EMAIL
    # this worked without pin_digests... but resulted in failing pipeline upstream
    if [ ! -e workspace-template.yml ]; then
       echo "missing workspace-template.yml"
       exit 1
    fi
    $TKN pipeline start operator-ci-pipeline \
        --use-param-defaults \
        --param git_repo_url=${GIT_FORK_REPO_URL} \
        --param git_branch=${GIT_BRANCH} \
        --param upstream_repo_name=${UPSTREAM_REPO_NAME} \
        --param bundle_path=${OPERATOR_BUNDLE_PATH} \
        --param env=prod \
        --workspace name=pipeline,volumeClaimTemplateFile=workspace-template.yml \
        --workspace name=ssh-dir,secret=github-ssh-credentials \
        --showlog \
        --param git_username=${GIT_USERNAME} \
        --param git_email=${GIT_EMAIL} \
        --param tlsverify=false \
        --param pin_digests=true \
        --param submit=${SUBMIT}
}


# convenience aggregator functions:

create_cluster_infra(){
    create_pipeline_project
    create_kubeconfig_secret
    import_catalogs
}

install_pipeline_manual(){
    fetch_certification_pipeline
    install_certification_pipeline
}

install_pipeline_automated(){
    # depends on Operator Certification Operator being installed
    _must_have OCP_PROJECT KUBECONFIG
    kubectl apply -f - << EOPIPELINE
apiVersion: certification.redhat.com/v1alpha1
kind: OperatorPipeline
metadata:
  name: certification-pipeline
  namespace: ${OCP_PROJECT}
spec:
  applyCIPipeline: true
  applyHostedPipeline: false
  applyReleasePipeline: false
  gitHubSecretName: github-api-token
  kubeconfigSecretName: kubeconfig
  operatorPipelinesRelease: main
  pyxisSecretName: pyxis-api-secret
EOPIPELINE
}

create_workspace_template(){
    cat > workspace-template.yml << EOWTEMPLATE
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 5Gi
EOWTEMPLATE
}

cleanup_secrets(){
    _must_have GITHUB_TOKEN_FILE PYXIS_API_KEY_FILE SSH_KEY_FILE
    $OC delete secret github-api-token
    $OC delete secret pyxis-api-secret
    # Needed during digest pinning
    $OC delete secret github-ssh-credentials
}

setup_and_run(){
    create_cluster_infra
    install_pipeline
    create_secrets
    run_certification_pipeline
}

if [ $# -lt 1 ]; then
    setup_and_run
fi

for cmd in "$@"
do
    $cmd
done