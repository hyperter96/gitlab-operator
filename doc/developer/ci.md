# CI

## OpenShift CI clusters

We have created OpenShift clusters in GKE that are used for acceptance tests, including QA suite.

kubeconfig files for connecting to these clusters are stored in 1Password, cloud-native vault.
Search for `ocp-ci`.

CI clusters have been launched with `scripts/create_openshift_cluster.sh` in this project. CI variables named `KUBECONFIG_OCP_4_7` allow scripts to connect to clusters as kubeadmin. `4_7` refers to the major and minor version of the targeted OpenShift cluster.

See [doc/doc/openshift-cluster-setup.md](../doc/openshift-cluster-setup.md) for instruction on using this script.
