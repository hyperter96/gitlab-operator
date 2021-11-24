# Troubleshooting the Operator

This document is a collection of notes and tips to assist in troubleshooting
the installation of the GitLab Operator and the deployment of a GitLab
instance from the GitLab custom resource.

## Installation problems

Troubleshooting the installation of the operator in a Kubernetes environment
is much like troubleshooting any other Kubernetes workload. After deploying
the operator manifest monitor the output of `kubectl describe` for the
operator Pod or `kubectl get events -n <namespace>`. This will indicate any
problems with retrieving the operator image or any other pre-condition for
starting the operator.

If the operator is starting up, but exits prematurely, examining the operator
logs can provide information for determining the cause for the Pod's
termination. This can be done with the following command:

```shell
kubectl logs deployment/gitlab-controller-manager -c manager -f -n <namespace>
```

Additionally, the operator depends on Cert Manager in order to create TLS
certificate for proper operation. The TLS certificate gets created as a
Secret and mounted as a volume on the operator Pod. Problems with obtaining
the TLS certificate can be found in the event log for the Namespace.

```shell
$ kubectl get events -n gitlab-system
...
102s        Warning   FailedMount         pod/gitlab-controller-manager-d4f65f856-b4mdj    MountVolume.SetUp failed for volume "cert" : secret "webhook-server-cert" not found
107s        Warning   FailedMount         pod/gitlab-controller-manager-d4f65f856-b4mdj    Unable to attach or mount volumes: unmounted volumes=[cert], unattached volumes=[cert gitlab-manager-token-fc4p9]: timed out waiting for the condition
...
```

The next step would be to inspect the Cert Manager logs looking for issues
that indicate the failure in creating the TLS certificate.

### OpenShift specific problems

OpenShift has a more restrictive security model and as a result, the GitLab
operator needs to be installed with the cluster administrator account. The
developer accounts do not have the necessary privileges to allow the operator
function properly.

## Problems with deployment of GitLab instance

In addition to the information presented here, one should consult the
GitLab Helm chart [troubleshooting documentation](https://docs.gitlab.com/charts/troubleshooting/index.html).

- [Core services not ready](#core-services-not-ready)
- [NGINX Ingress Controller pods missing](#nginx-ingress-controller-pods-missing)
- [Horizontal pod autoscalers are not scaling](#horizontal-pod-autoscalers-are-not-scaling)
- [Ingress does not show external IP](#ingress-does-not-show-external-ip)

### Core services not ready

The GitLab Operator relies on installing instances of Redis, PostgreSQL and
Gitaly. These are known as the core services. If after deploying a GitLab
customer resource there are an excessive number of operator log messages
stating that the core services are not ready, then it is one of these
services that is having problems becoming operational.

Specifically check the endpoints for each of these services to insure that
they are getting connected to the service's Pod. This is also a possible
indication that the cluster does not have enough resources to support the
GitLab instance and additional nodes should be added to the cluster.

Issue [#305](https://gitlab.com/gitlab-org/cloud-native/gitlab-operator/-/issues/305)
has been created to track the reporting of which core service is stopping
the deployment of the GitLab instance.

### NGINX Ingress Controller pods missing

In an OpenShift environment the
[NGINX Ingress Controller](https://kubernetes.github.io/ingress-nginx)
is used in place of OpenShift Routes for directing traffic to the GitLab
instance (both HTTPS and SSH). If you are having a problem with connecting
to the GitLab instance, first insure that there is a deployment for the
NGINX Ingress Controller.

If a deployment is present, check the `READY` column of the
`kubectl get deploy` output. If the `READY` status is reported back as
`0/0`, then inspect the output of
`kubectl get events -n <namespace> | grep -i nginx` looking for messages
that state that the Security Context Constraint (SCC) has been violated.

This is an indication that the NGINX RBAC resources for OpenShift were
not deployed. The operator manifest for OpenShift should be reapplied with
the following command:

```shell
kubectl apply -f https://gitlab.com/api/v4/projects/18899486/packages/generic/gitlab-operator/<VERSION>/gitlab-operator-openshift.yaml
```

After the manifest has been applied, it may be necessary to delete the
Ingress controller Deployment to acquire the SCC properly and allow the
Ingress controller to create the Pods correctly.

### Horizontal pod autoscalers are not scaling

If it is found that the horizontal pod autoscalers (HPA) do not scale the
number of pods according to traffic load, then check for an installation
of the Metrics Server. In a Kubernetes cluster the Metrics Server is an
additional component that needs to be installed. The installation process
can be found in the [installation documentation](installation.md#metrics).

An OpenShift cluster has a built in Metrics Server and as a result the
HPAs should operate correctly.

### Ingress does not show external IP

In an OpenShift environment the NGINX Ingress may receive the hostname of
the GitLab instance instead of the external IP address. This can be seen in
the output of `kubectl get ingress -n <namespace>` in the `ADDRESS` column.

This is caused by the OpenShift Router controller updating the Ingress resource
instead of ignoring it because it is a different Ingress class. The following
command will instruct the OpenShift Router controller to ignore Ingresses other
than the standard Ingresses deployed in OpenShift:

```shell
  kubectl -n openshift-ingress-operator \
    patch ingresscontroller default \
    --type merge \
    -p '{"spec":{"namespaceSelector":{"matchLabels":{"openshift.io/cluster-monitoring":"true"}}}}'
```
