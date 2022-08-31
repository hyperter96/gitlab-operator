# RedHat Operator Bundle certification process

This document outlines certification process for OLM bundle submission for RedHat Marketplace. It is based on [Red Hat Software Certification Workflow Guide](https://access.redhat.com/documentation/en-us/red_hat_software_certification/8.49/html/red_hat_software_certification_workflow_guide/assembly-running-the-certification-suite-locally_openshift-sw-cert-workflow-complete-pre-certification-checklist)

## Provision OpenShift cluster

Existing OpenShift cluster is required. If using already provisioned OpenShift cluster adjust steps accordingly.

### OpenShift-provisioning pipeline

One way to provision a new OpenShift cluster is using [OpenShift-provisioning pipeline](https://gitlab.com/gitlab-org/distribution/infrastructure/openshift-provisioning)

Pipeline creates convenient artifact that includes auth information for cluster as well as all necessary binaries. Download Zip file for the artifact produced by `deploy_cluster` job and extract it into some convenient location (`${HOME}/mycluster`)

```shell
BASEDIR=${HOME}/mycluster
OC=${BASEDIR}/bin/oc-x.y.z
TKN=${BASEDIR}/bin/tkn
KUBECONFIG=${BASEDIR}/x.y.z/auth/kubeconfig
export OC TKN KUBECONFIG
```

### Manual provisioning

Follow OpenShift installation documentation to provision new cluster. Make sure to extract and preserve:

- `oc` binary corresponding to OCP release set up
- download `tkn` binary compatible with the release
- contents of the `auth` directory under the installation root directory

```shell
OC=/path/to/oc
TKN=/path/to/tkn
KUBECONFIG=/path/to/kubeconfig
export OC TKN KUBECONFIG
```

## Setup OpenShift cluster

```shell
export KUBECONFIG
redhat/operator-certification/scripts/operator_certification_pipeline.sh create_cluster_infra
```

## Setup repo

### Create API token (PAT)

In GitHub navigate to [profile settings](https://github.com/settings/profile) `Developer settings`/`Personal access tokens` and generate a new one with scope `repo`. Save it to a local file in secure location (`${HOME}/secure/github_api_token.txt`)

Note that this token as access to **all** of your GitHub repos at this point.

### fork repo

You'll need to fork one of:

- [certified-operators](https://github.com/redhat-openshift-ecosystem/certified-operators) or
- [redhat-marketplace-operators](https://github.com/redhat-openshift-ecosystem/redhat-marketplace-operators)

for example:

1. Using GitHub UI create a fork
1. clone fork locally:

   ```shell
   VERSION="0.9.1" # Version to be submitted

   cd
   git clone git@github.com:<YOUR-GITHUB-ACCOUNT>/certified-operators.git
   git checkout -b gitlab-operator-kubernetes-${VERSION}
   CATALOG_REPO_CLONE=${HOME}/certified-operators
   ```

### Create Deploy Keys

create SSH key pair with empty passphrase:

```shell
ssh-keygen -f certified-operators-key
```

for that repo create `Deploy Keys` (`Settings`/`Deploy keys`/`Add new` for the repo)

and add `certified-operators-key.pub`

## Setup pipeline

### Pre-requisites

- [Provisioned OpenShift cluster](#provision-openshift-cluster) you should have:
  - associated `kubeconfig` file (`$KUBECONFIG`)
  - associated `tkn` binary (`$TKN`)
  - associated `oc` binary (`$OC`)
- GitHub PAT (`$GITHUB_TOKEN_FILE`)
- Pyxis API Token (`$PYXIS_API_KEY_FILE`) obtained from RedHat
- SSH Key pair (`$SSH_KEY_FILE` - private key file)
  - Added as "deployment key" to forked project

**NOTE** make sure files `$GITHUB_TOKEN_FILE` and `$PYXIS_API_KEY_FILE` do not contain newline character (`0x0a` at the end of the file):

```shell
hexdump -C $GITHUB_TOKEN_FILE
hexdump -C $PYXIS_API_KEY_FILE
```

then execute create secrets and install pipeline:

```shell
GITHUB_TOKEN_FILE=/path/to/github_token.txt \
  PYXIS_API_KEY_FILE=/path/to/pyxis_api_key.txt \
  SSH_KEY_FILE=/path/to/ssh_private_key \
  KUBECONFIG="${BASEDIR}/auth/kubeconfig" \
  redhat/operator-certification/scripts/operator_certification_pipeline.sh create_secrets install_pipeline
```

## Generate bundle

```shell
VERSION="0.9.1" # Version to be submitted

OSDK_BASE_DIR=".build/cert" \
    DOCKER="podman" \
    OLM_PACKAGE_VERSION=${VERSION} \
    OPERATOR_TAG=${VERSION} \
    scripts/olm_bundle.sh build_manifests generate_bundle patch_bundle
```

## Properly annotate bundle for submission

```shell
BUNDLE_DIR=.build/cert/bundle \
    redhat/operator-certification/scripts/configure_bundle.sh adjust_annotations adjust_csv
``` 

## Copy & Push changes into the forked repo

At this point one needs to copy bundle to it's new location (you'll need value of `CATALOG_REPO_CLONE` from [fork repo](#fork-repo)):

```shell
VERSION="0.9.1" # Version to be submitted

cp -r .build/cert/bundle ${CATALOG_REPO_CLONE}/operators/gitlab-operator-kubernetes/${VERSION}
( cd ${CATALOG_REPO_CLONE} && git add operators/gitlab-operator-kubernetes/${VERSION} \ 
   && git commit -am "Add gitlab-operator-${VERSION}" \
   && git push origin gitlab-operator-kubernetes-${VERSION})
```

## Run certification pipeline

GitHub Username and email will need to be obtained for this step and used respectively in `GIT_USERNAME` and `GIT_EMAIL`

```shell
VERSION="0.9.1" # Version to be submitted

GIT_USERNAME="<developer_gh_username>" \
  GIT_EMAIL="<developer_gh_email>" \
  OPERATOR_BUNDLE_PATH="operators/gitlab-operator-kubernetes/${VERSION}" \
  KUBECONFIG="${BASEDIR}/auth/kubeconfig" \
  GIT_FORK_REPO_URL="git@github.com:<developer_gh_username>/certified-operators.git" \
  GIT_BRANCH="gitlab-operator-kubernetes-${VERSION}" \
  redhat/operator-certification/scripts/operator_certification_pipeline.sh run_certification_pipeline
```

this will create upstream PR and open submission in RH portal

**NOTE** if you are getting 

```plaintext
ValueError: Invalid header value b'Bearer XXXXXXXXXXXXXXXXXXXXXXX\n'
```

in the output - likely one of the Secrets contains unwanted newline: `pyxis-api-secret` or `github-api-token` context of the error should help determine which one.
