# CI

## OpenShift CI clusters

We have created OpenShift clusters in GKE that are used for acceptance tests, including QA suite.

kubeconfig files for connecting to these clusters are stored in 1Password, cloud-native vault.
Search for `ocp-ci`.

CI clusters have been launched with `scripts/create_openshift_cluster.sh` in this project. CI variables named `KUBECONFIG_OCP_4_7` allow scripts to connect to clusters as kubeadmin. `4_7` refers to the major and minor version of the targeted OpenShift cluster.

See [doc/doc/openshift-cluster-setup.md](../doc/openshift-cluster-setup.md) for instruction on using this script.

### Scaling OpenShift clusters

OpenShift clusters use `machineset`s to pool resources. Default deployment has 4 `machineset`s out of which 2 are utilized. 
We retain that approach and scale 2 active `machineset`s to 2 nodes.

Following [OpenShift scaling documentation](https://docs.openshift.com/container-platform/4.7/scalability_and_performance/recommended-cluster-scaling-practices.html) :

```shell
$ export KUBECONFIG=~/work/ocp/kubeconfig_4_7 
$ oc get machinesets -n openshift-machine-api
NAME                         DESIRED   CURRENT   READY   AVAILABLE   AGE
ocp-ci-4717-abcde-worker-a   1         1         1       1           59d
ocp-ci-4717-abcde-worker-b   1         1         1       1           59d
ocp-ci-4717-abcde-worker-c   0         0                             59d
ocp-ci-4717-abcde-worker-f   0         0                             59d
```

create temporary capacity to handle workloads while we're scaling down:

```shell
$ oc scale --replicas=1 machineset ocp-ci-4717-abcde-worker-c -n openshift-machine-api
machineset.machine.openshift.io/ocp-ci-4717-abcde-worker-c scaled
$ oc get machinesets -n openshift-machine-api
NAME                         DESIRED   CURRENT   READY   AVAILABLE   AGE
ocp-ci-4717-abcde-worker-a   1         1         1       1           59d
ocp-ci-4717-abcde-worker-b   1         1         1       1           59d
ocp-ci-4717-abcde-worker-c   1         1                             59d
ocp-ci-4717-abcde-worker-f   0         0                             59d
$ kubectl get nodes
NAME                                                              STATUS     ROLES    AGE   VERSION
ocp-ci-4717-abcde-master-0.c.cloud-native-123456.internal         Ready      master   59d   v1.20.0+2817867
ocp-ci-4717-abcde-master-1.c.cloud-native-123456.internal         Ready      master   59d   v1.20.0+2817867
ocp-ci-4717-abcde-master-2.c.cloud-native-123456.internal         Ready      master   59d   v1.20.0+2817867
ocp-ci-4717-abcde-worker-a-wz8ts.c.cloud-native-123456.internal   Ready      worker   59d   v1.20.0+2817867
ocp-ci-4717-abcde-worker-b-shnf4.c.cloud-native-123456.internal   Ready      worker   59d   v1.20.0+2817867
ocp-ci-4717-abcde-worker-c-5gq6r.c.cloud-native-123456.internal   NotReady   worker   15s   v1.20.0+2817867
```

Scale down one of the `machineset`s:

```shell
$ oc scale --replicas=0 machineset ocp-ci-4717-abcde-worker-a -n openshift-machine-api
```

Wait for scale down to complete:

```shell
$ oc scale --replicas=2 machineset ocp-ci-4717-abcde-worker-a -n openshift-machine-api
$ kubectl get nodes
NAME                                                              STATUS   ROLES    AGE     VERSION
ocp-ci-4717-abcde-master-0.c.cloud-native-123456.internal         Ready    master   59d     v1.20.0+2817867
ocp-ci-4717-abcde-master-1.c.cloud-native-123456.internal         Ready    master   59d     v1.20.0+2817867
ocp-ci-4717-abcde-master-2.c.cloud-native-123456.internal         Ready    master   59d     v1.20.0+2817867
ocp-ci-4717-abcde-worker-b-shnf4.c.cloud-native-123456.internal   Ready    worker   59d     v1.20.0+2817867
ocp-ci-4717-abcde-worker-c-5gq6r.c.cloud-native-123456.internal   Ready    worker   7m18s   v1.20.0+2817867
```

scale back up to needed number of nodes:

```shell
$ oc scale --replicas=2 machineset ocp-ci-4717-abcde-worker-a -n openshift-machine-api
machineset.machine.openshift.io/ocp-ci-4717-abcde-worker-a scaled
$ kubectl get nodes
NAME                                                              STATUS   ROLES    AGE   VERSION
ocp-ci-4717-abcde-master-0.c.cloud-native-123456.internal         Ready    master   59d   v1.20.0+2817867
ocp-ci-4717-abcde-master-1.c.cloud-native-123456.internal         Ready    master   59d   v1.20.0+2817867
ocp-ci-4717-abcde-master-2.c.cloud-native-123456.internal         Ready    master   59d   v1.20.0+2817867
ocp-ci-4717-abcde-worker-a-n7qvs.c.cloud-native-123456.internal   Ready    worker   72s   v1.20.0+2817867
ocp-ci-4717-abcde-worker-a-sqn77.c.cloud-native-123456.internal   Ready    worker   73s   v1.20.0+2817867
ocp-ci-4717-abcde-worker-b-shnf4.c.cloud-native-123456.internal   Ready    worker   59d   v1.20.0+2817867
ocp-ci-4717-abcde-worker-c-5gq6r.c.cloud-native-123456.internal   Ready    worker   11m   v1.20.0+2817867
```

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
