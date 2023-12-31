---
stage: Systems
group: Distribution
info: To determine the technical writer assigned to the Stage/Group associated with this page, see https://about.gitlab.com/handbook/product/ux/technical-writing/#assignments
---

# Test OLM Bundles

## Pre-requisites

Versions listed are the ones known to work, lower versions may work but were not tested

- `task-3.17.0` (`asdf`)
- `operator-sdk-1.32.0` (`asdf`)
- `kubectl-1.25.3` (`asdf`)
- `helm-3.10.1` (`asdf`)
- `kustomize-4.5.7` (`asdf`)
- `yq-4.29.2` (`asdf`)
- `opm-1.26.2` (is auto-downloaded by script, or can be installed via `asdf` using [asdf-opm](https://gitlab.com/dmakovey/asdf-opm.git) plugin)
- `kind-0.17.0` (`asdf`)
- `docker` (could be replaced by `podman` via `DOCKER="podman"`)
- `podman` (some of the OperatorSDK toolchain use podman)
- `k9s-0.26.7` (`asdf` **OPTIONAL**)

## Set up environment

### Set up Podman

```shell
podman login registry.gitlab.com
```

### Set up Docker

If you're not using `podman` for all tasks, authorize Docker to access `registry.gitlab.com`:

```shell
docker login registry.gitlab.com
```

### Set up Git

Ensure `user.name` and `user.email` are configured in Git.

### Set up environment

|         variable name          | required |                               default                                |                                                                 description                                                                 |
| ------------------------------ | -------- | -------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| `OPERATOR_VERSION`             | yes      |                                                                      | Version of Operator to upgrade to                                                                                                           |
| `PREVIOUS_OPERATOR_VERSION`    | yes      |                                                                      | Version of Operator to upgrade from                                                                                                         |
| `LOCAL_IP`                     | yes      | `127.0.0.1`                                                          | Local machine IP, needed for `GITLAB_OPERATOR_DOMAIN`. When `GITLAB_OPERATOR_DOMAIN` is provided - `LOCAL_IP` can be omitted                |
| `GITLAB_OPERATOR_OLM_REGISTRY` | yes      | `registry.gitlab.com/gitlab-org/cloud-native/gitlab-operator/bundle` | OLM Bundles and Catalogs Registry with released bundles and catalogs                                                                        |
| `BUNDLE_REGISTRY`              | yes      |                                                                      | Staging container registry to publish intermediary OLM Bundles and Catalogs to                                                              |
| `OLM_TESTING_ENVIRONMENT`      | no       | `./test_olm.env`                                                     | File containing environment variables necessary for test runs                                                                               |
| `OPERATOR_TAG`                 | no       | `$OPERATOR_VERSION`                                                  | Operator Container tag to test                                                                                                              |
| `OLM_BUNDLE_SH`                | no       | `scripts/olm_bundle.sh`                                              | Path to underlying `olm_bundle.sh` script                                                                                                   |
| `PROVISION_AND_DEPLOY_SH`      | no       | `scripts/provision_and_deploy.sh`                                    | Path to underlying `provision_and_deploy.sh` script                                                                                         |
| `YQ`                           | no       | `yq`                                                                 | Path to `yq` utility                                                                                                                        |
| `OPM_VERSION`                  | no       | `1.26.2`                                                             | `opm` version to automatically fetch if no binary specified in `OPM`                                                                        |
| `OPM`                          | no       | `.build/opm`                                                         | Path to `opm` binary. Auto-fetched if empty (using `OPM_VERSION` )                                                                          |
| `OSDK_BASE_DIR`                | no       | `.build/operatorhub-io`                                              | Directory for intermediate OLM Bundling artifacts storage                                                                                   |
| `OLM_PACKAGE_VERSION`          | no       | `$OPERATOR_TAG`                                                      | Version to apply to OLM package                                                                                                             |
| `KUBERNETES_TIMEOUT`           | no       | `120s`                                                               | Timeout for Kubernetes operations                                                                                                           |
| `DO_NOT_PUBLISH`               | no       | `""`                                                                 | Controls whether to compile and publish current bundle/catalog (to a temporary registry) or use already published ones from public registry |
| `BUNDLE_VERSION`               | no       | `$OPERATOR_VERSION`                                                  | Version of the bundle to upgrade to                                                                                                         |
| `PREVIOUS_BUNDLE_VERSION`      | no       | `$PREVIOUS_OPERATOR_VERSION`                                         | Version of the bundle to upgrade from                                                                                                       |
| `PREVIOUS_CHART_VERSION`       | no       | autogenerated                                                        | Chart version to upgrade from                                                                                                               |
| `GITLAB_OPERATOR_DOMAIN`       | no       | `${LOCAL_IP}.nip.io`                                                 | Domain to deploy test GitLab instance to                                                                                                    |
| `GITLAB_OPERATOR_DIR`          | no       | `.`                                                                  | Directory with GitLab Operator repository                                                                                                   |
| `GITLAB_CHART_VERSION`         | no       | first line in `${GITLAB_OPERATOR_DIR}/CHART_VERSIONS}`               | Chart Version to upgrade to                                                                                                                 |
| `GITLAB_CHART_REPO`            | no       | `https://gitlab.com/gitlab-org/charts/gitlab`                        | GitLab Helm Chart repository HTTP URI. Mainly used to fetch default KinD configs.                                                           |
| `K8S_VERSION`                  | no       | `1.22.4`                                                             | K8s version to use for cluster setup                                                                                                        |
| `KIND_CONFIG`                  | no       | `examples/kind/kind-ssl.yaml` from GitLab Chart's default branch     | KinD configuration file to prepare KinD cluster for GitLab deployment                                                                       |

For additional variables look at [Provision and deploy](provision_and_deploy.md) and [OperatorHub publishing](operatorhub_publishing.md)

Create `test_olm.env` in Operator's root dir (or point to a file in a custom location using `${OLM_TESTING_ENVIRONMENT}` environment variable)

Make sure to confirm every following customization line reflects **your personal** setup:

```shell
export OPERATOR_VERSION="0.10.2"
export BUNDLE_REGISTRY="registry.gitlab.com/dmakovey/gitlab-operator-bundle"

export PREVIOUS_OPERATOR_VERSION="0.9.1"
export LOCAL_IP="192.168.3.194"
export DO_NOT_PUBLISH="yes" # do not re-compile and publish bundle/catalog
                            # use the ones already published
```

## Prepare KinD Cluster (Optional)

To run tests in KinD cluster with OLM set up:

```shell
scripts/test_olm.sh setup_kind_cluster
```

Otherwise current `kubectl` context should point to an existing cluster that has OLM pre-installed

## Running tests

1. Setup KinD cluster and deploy Operator there:

   ```shell
   scripts/test_olm.sh upgrade_test_step1
   ```

   Operator Pod should be up and running. Confirm by running:

   ```shell
   kubectl get pod -n gitlab-system -l control-plane=controller-manager
   ```

1. Deploy "old" version of GitLab via Operator:

   ```shell
   scripts/test_olm.sh upgrade_test_step2
   ```

   Confirm GitLab has been deployed:

   ```shell
   kubectl get -n gitlab-system gitlab
   ```

1. Upgrade the *Operator*:

   ```shell
   scripts/test_olm.sh upgrade_test_step3
   ```

   Confirm that Operator has been upgraded.

   **Wait** for the install to roll out before next step

1. Upgrade GitLab:

   ```shell
   scripts/test_olm.sh upgrade_test_step4
   ```

   Confirm that GitLab has been upgraded and is functional

   **Wait** for the upgrade to complete

1. Confirm GitLab is running:

   ```shell
   scripts/test_olm.sh check_gitlab
   ```

   above will query Operator for GitLab status, alternatively use

   ```shell
   scripts/test_olm.sh check_gitlab2
   ```

   which bypasses Operator checks and runs checks against GitLab instance itself.
