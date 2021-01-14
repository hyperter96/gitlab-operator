# Installing operator from source

This document describes how to deploy the GitLab operator via manifests in your Kubernetes or Openshift cluster.

These steps normally are handled by OLM, the Operator Lifecycle Manager, once an operator is bundle published. However, to test the most recent operator images, users may need to install the operator using the deployment manifests available in the operator repository.

## Requirements

0. Create an OpenShift cluster, see [openshift-cluster-setup.md](openshift-cluster-setup.md).

1. Ensure the operators it depends on are present. These operators should be installed via the in-cluster OperatorHub.

    The GitLab operator uses the following operators:

    * the `Nginx Ingress Operator` by Nginx Inc. to deploy and Ingress Controller. This should be deployed from operatorhub.io if using Kubernetes or the embedded Operator Hub on Openshift environments

    * the `Cert Manager operator` to create certificates used to secure the GitLab and Registry urls. Once this operator has been installed, create a cert-manager instance. Use default "cert-manager" for the Name field, the Labels field can be blank.

2. Clone the GitLab operator repository to your local system

    ```
    $ git clone https://gitlab.com/gitlab-org/gl-openshift/gitlab-operator.git
    $ cd gitlab-operator
    ```

3. Deploy the CRDs(Custom Resource Definitions) for the resources managed by the operator

    ```
    $ make install
    ```

4. Deploy the operator

    ```
    $ make deploy
    ```

    This command first deploys the service accounts, roles and role bindings used by the operator, and then the operator itself.
