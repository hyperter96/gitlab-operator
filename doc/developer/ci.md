# CI

## OpenShift CI clusters

We have created OpenShift clusters in GKE that are used for acceptance tests, including QA suite.

kubeconfig files for connecting to these clusters are stored in 1Password, cloud-native vault.
Search for `ocp-ci`.

CI clusters have been launched with `scripts/create_openshift_cluster.sh` in this project. CI variables named `KUBECONFIG_OCP_4_7` allow scripts to connect to clusters as kubeadmin. `4_7` refers to the major and minor version of the targeted OpenShift cluster.

See [doc/doc/openshift-cluster-setup.md](../doc/openshift-cluster-setup.md) for instruction on using this script.

### external-dns

[external-dns](https://github.com/kubernetes-sigs/external-dns) has been installed using the [Bitnami Helm Chart](https://github.com/bitnami/charts/tree/master/bitnami/external-dns) using [ci/scripts/install_external_dns.sh](../ci/scripts/install_external_dns.sh). This tool creates DNS entries for the NGINX Ingress controller Services that are created as external-facing LoadBalancers, ensuring that our QA jobs can reach the instance for testing.

- For our `4.6` OpenShift Cluster: `CLUSTER_VERSION=4.6 ENVIRONMENT=openshift GOOGLE_APPLICATION_CREDENTIALS=gitlab-operator-ci-gcloud-externaldns.json ./ci/scripts/install_external_dns.sh`
- For our `4.7` OpenShift Cluster: `CLUSTER_VERSION=4.7 ENVIRONMENT=openshift GOOGLE_APPLICATION_CREDENTIALS=gitlab-operator-ci-gcloud-externaldns.json ./ci/scripts/install_external_dns.sh`

Note: `gitlab-operator-ci-gcloud-externaldns.json` is a file containing the credentials for the external-dns ServiceAccount created in GCP. You can find this credentials file in 1Password by searching for `externaldns` in the `Cloud Native` vault.

## Configuration

### Job timeouts

Note: timeouts for Jobs can be configured. If the timeout is reached, then the GitLab Controller will return an error that the Job could not be completed in time.

To configure these, modify the values under `spec.template.spec.containers[0].env` in
[config/manager/manager.yaml](../../config/manager/manager.yaml).
