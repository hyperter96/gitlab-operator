# Installation

This document describes how to deploy the GitLab operator via manifests in your Kubernetes or OpenShift cluster.

If using OpenShift, these steps normally are handled by OLM (the Operator Lifecycle Manager) once an operator is bundle published. However, to test the most recent operator images, users may need to install the operator using the deployment manifests available in the operator repository.

## Prerequisites

Please consult the "Prerequisites" section of [installation.md](../installation.md#prerequisites).

## Installing the GitLab Operator

1. Clone the GitLab operator repository to your local system.

    ```
    $ git clone https://gitlab.com/gitlab-org/cloud-native/gitlab-operator.git
    $ cd gitlab-operator
    ```

2. Build CRDs and Operator manifests:

    ```
    $ make build_operator
    ```

   Note: in some cases, you may run into issues resolving dependencies and see an error message such as:

    ```
    go get: github.com/openshift/api@v3.9.0+incompatible: invalid version: unknown revision v3.9.0
    ```

   To address this, configure `GOPROXY` as mentioned [in this issue](https://github.com/openshift/api/issues/456#issuecomment-576842590):

    ```bash
    export GOPROXY="https://proxy.golang.org/"
    ```

3. Deploy the GitLab Operator.

    ```
    $ kubectl create namespace gitlab-system
    $ make deploy_operator
    ```

    This command first deploys the service accounts, roles and role bindings used by the operator, and then the operator itself.

    Note: by default, the Operator will only watch the namespace where it is deployed. If you would like it to watch at the cluster scope,
    modify [config/manager/kustomization.yaml](../config/manager/kustomization.yaml) by commenting out the `namesapce_scope.yaml` patch.

4. Create a GitLab custom resource (CR).

   Create a new file named something like `mygitlab.yaml`.

   Here is an example of the content to put in this file:

   ```yaml
   apiVersion: apps.gitlab.com/v1beta1
   kind: GitLab
   metadata:
     name: example
   spec:
     chart:
       version: "X.Y.Z" # select a version from the CHART_VERSIONS file in the root of this project
       values:
         global:
           hosts:
             domain: example.com # use a real domain here
           ingress:
             configureCertmanager: true
         certmanager-issuer:
           email: youremail@example.com # use your real email address here
   ```

   For more details on configuration options to use under `spec.chart.values`, see our [GitLab Helm Chart documentation](https://docs.gitlab.com/charts).

5. Deploy a GitLab instance using your new GitLab CR.

   ```
   $ kubectl -n gitlab-system apply -f mygitlab.yaml
   ```

   This command sends your GitLab CR up to the cluster for the GitLab Operator to reconcile. You can watch the progress by tailing the logs from the controller pod:

   ```
   $  kubectl -n gitlab-system logs deployment/gitlab-controller-manager -c manager -f
   ```

   You can also list GitLab resources and check their status:

   ```
   $ kubectl get gitlabs -n gitlab-system
   ```

   When the CR is reconciled (the status of the GitLab resource will be `RUNNING`), you can access GitLab in your browser at `https://gitlab.example.com`.

## Cleanup

Certain operations like file removal under `config/` directory may not trigger rebuild/redeploy, in which cases one should employ:

```shell
make clean
```

This will remove all of the build artifacts and the install record.

## Uninstall the GitLab Operator

Follow the steps below to remove the GitLab Operator and its associated resources.

Items to note prior to uninstalling the operator:

- The operator does not delete the Persistent Volume Claims or Secrets when a GitLab instance is deleted.
- When deleting the Operator, the namespace where it is installed (`gitlab-system` by default) will not be deleted automatically. This is to ensure persistent volumes are not lost unintentionally.

### Uninstall an instance of GitLab

```
$ kubectl -n gitlab-system delete -f mygitlab.yaml
```

This will remove the GitLab instance, and all associated objects except for (PVCs as noted above).

### Uninstall the GitLab Operator

```
$ make delete_operator
```

This will delete the Operator's resources, including the running Deployment. It will not delete objects associated with a GitLab instance.

### Uninstall CRDs

```
$ make uninstall_crds
```