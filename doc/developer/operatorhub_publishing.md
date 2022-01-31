# How to publish to OperatorHub.io

OperatorHub.io is a home for the Kubernetes community to share Operators.
To publish the GitLab Operator to OperatorHub:

1. Fork the [community-operators repository](https://github.com/k8s-operatorhub/community-operators).
1. Clone the forked community-operators repository:
   - If the fork was freshly created:

     ```shell
     git clone -o mine git@github.com:<your_github_username>/community-operators.git
     cd community-operators
     git remote add -o upstream https://github.com/k8s-operatorhub/community-operators.git
     ```

   - If this is a subsequent update to the already created fork:

     ```shell
     cd community-operators
     git fetch --all
     git checkout main
     git rebase -i upstream/main
     ```

1. Set up your shell:

   ```shell
   # Published Operator Image tag
   export OPERATOR_TAG="0.3.1"
   # Version that we're going to apply to OLM
   export OLM_PACKAGE_VERSION=${OPERATOR_TAG}
   export OPERATORHUB_DIR="${HOME}/work/community-operators"
   export OPERATORHUB_NAME="gitlab-operator-kubernetes"
   export OSDK_BASE_DIR=".build/operatorhub-io"
   export OPERATOR_SDK="${HOME}/bin/operator-sdk_linux_amd64"
   ```

1. Create a new branch:

   ```shell
   cd ${OPERATORHUB_DIR}
   git checkout -B gitlab-release-${OPERATOR_TAG}
   ```

1. Edit [`config/manifests/bases/gitlab-operator-kubernetes.clusterserviceversion.yaml`](https://gitlab.com/gitlab-org/cloud-native/gitlab-operator/-/blob/master/config/manifests/bases/gitlab-operator-kubernetes.clusterserviceversion.yaml)
   under `metadata.annotations.alm-examples` to refer to a valid chart version
   that this version of Operator ships with.

1. Create the Operator bundle and copy the files to the right location:

   ```shell
   # assemble bundle
   scripts/olm_bundle.sh build_manifests generate_bundle patch_bundle
   # validate bundle
   scripts/olm_bundle.sh validate_bundle

   mkdir -p ${OPERATORHUB_DIR}/operators/${OPERATORHUB_NAME}/${OLM_PACKAGE_VERSION}
   cp -r ${OSDK_BASE_DIR}/bundle/* ${OPERATORHUB_DIR}/operators/${OPERATORHUB_NAME}/${OLM_PACKAGE_VERSION}
   ```

1. Add and commit (with sign) your changes:

   ```shell
   cd ${OPERATORHUB_DIR}
   git add operators/${OPERATOR_HUB_NAME}/${OLM_PACKAGE_VERSION}
   git commit -s
   ```

1. Push your branch to your fork and create a pull request upstream.
   Wait for approval from GitLab team members and/or OperatorHub reviewers
   before the merge is completed.
