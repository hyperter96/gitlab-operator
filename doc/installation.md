# Installing operator from source

This document describes how to deploy the GitLab operator via manifests in your Kubernetes or Openshift cluster.

These steps normally are handled by OLM, the Operator Lifecycle Manager, once an operator is bundle published. However, to test the most recent operator images, users may need to install the operator using the deployment manifests available in the operator repository.

## Requirements

0. Create an OpenShift cluster, see [openshift-cluster-setup.md](openshift-cluster-setup.md).

1. Clone the GitLab operator repository to your local system

    ```
    $ git clone https://gitlab.com/gitlab-org/gl-openshift/gitlab-operator.git
    $ cd gitlab-operator
    ```

2. Ensure the operators it depends on are present. These operators can be installed via the in-cluster OperatorHub or via Make:

   ```
   $ make install_required_operators
   ```

   The GitLab operator uses the following operators:
   * the `Nginx Ingress Operator` by Nginx Inc. to deploy and Ingress Controller. This should be deployed from operatorhub.io if using Kubernetes or the embedded Operator Hub on OpenShift environments

   * the `Cert Manager operator` to create certificates used to secure the GitLab and Registry urls. Once this operator has been installed, create a cert-manager instance. Use default "cert-manager" for the Name field, the Labels field can be blank.


3. Deploy the CRDs(Custom Resource Definitions) for the resources managed by the operator

    ```
    $ make install_crds
    ```

4. Deploy the operator

    ```
    $ make deploy_operator
    ```

    This command first deploys the service accounts, roles and role bindings used by the operator, and then the operator itself.

5. Deploy a GitLab instance

   ```
   $ DOMAIN=mydomain.com make deploy_sample_cr
   ```

   This command injects the relevant values into `config/samples/apps_v1beta1_gitlab.yaml`, and then applies the custom resource.

6. Clean up

   The operator does not delete the persistent volume claims that hold the stateful data when a GitLab instance is deleted. Therefore, remember to delete any lingering volumes.

   When deleting the Operator, the namespace where it is installed (`gitlab-system` by default) will not be deleted automatically. This is to ensure persistent volumes are not lost unintentionally.

   ```
   $ make delete_sample_cr
   $ make delete_operator
   $ make uninstall_crds
   $ make uninstall_required_operators
   ```
