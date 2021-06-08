# Developer Guide

This developer guide aims to walk a new developer on how to setup up their environment to be able to contribute to this project.

## Setting up development environment

To setup your system for development of the operator, follow the steps below:

1. Install [golang](https://golang.org/dl/) in your environment.

   ```shell
   $ go version
   ```

2. Download `operator-sdk`. Current `operator-sdk` [releases](https://github.com/operator-framework/operator-sdk/releases) can be found in the projects repository.

   To check your version of operator SDK run,

   ```shell
   $ operator-sdk version`
   ```

   To contribute code to the current GitLab operator release, you will need at least operator SDK v1.0.0.

3. Clone the `gitlab-operator` repository into your GOPATH.

   ```shell
   $ git clone git@gitlab.com:gitlab-org/gl-openshift/gitlab-operator.git
   ```

## Project structure

The GitLab operator is built using the Operator SDK v1.0.0 and consequently uses the Kubebuilder v2 layout format. This is necessary to know since there was a change in project directory and some of the tooling used by operator SDK.

```
$ pwd
gitlab-operator
$ tree -dL 2 .
.
├── api
│   └── v1beta1
├── bundle
│   ├── manifests
│   ├── metadata
│   └── tests
├── config
│   ├── certmanager
│   ├── crd
│   ├── default
│   ├── deploy
│   ├── manager
│   ├── manifests
│   ├── prometheus
│   ├── rbac
│   ├── samples
│   ├── scorecard
|   ├── test
│   └── webhook
├── controllers
│   ├── backup
│   ├── gitlab
│   ├── helpers
│   ├── runner
│   ├── settings
│   ├── testdata
│   └── utils
├── doc
├── hack
│   └── assets
├── helm
│   └── testdata
└── scripts
    └── manifests
```

  * The `controllers` directory contains the controller implementations for the GitLab and GitLab Backup controllers.
  * The `api` directory contains the API resource definitions for the GitLab and GLBackup resources owned by the operator. The API definitions are grouped by their API version.
    The `*_types.go` file inside `api/<api_version>` contains spec definitions and markers used to generate the Custom Resource Definitions and Cluster Service Version file used by OLM.
  * The `config/samples` directory contains an example manifest for the GitLab Custom Resource.
  * The `config/test` directory contains a parametrized GitLab definition used for running integration tests.
  * The `config/rbac` directory contains the roles, role bindings, and service accounts needed for the operator to run. The roles should be updated through RBAC(Role Based Access Control) [markers](https://book.kubebuilder.io/reference/markers/rbac.html) inside your controllers.

    An example is shown below:

    `// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete`

    The contents of `config/rbac/custom` were created manually and is not affected by the RBAC markers.

    Most of the other contents of the config directory are automatically generated but could be modified using `kustomize`.

  * The `hack/assets` path contains resources that would need to be pushed inside the operator image when the container image is being built. This is where release files would go.

### Additional Resources


The `Makefile` allows us to customize manage different tasks such as:

 - Creating an Operator Lifecycle Manager bundle

   ```
   $ make bundle
   ```

 - Building a container image for the operator

   ```
   $ make docker-build IMG=quay.io/<username>/gitlab-operator:latest
   ```

 - Pushing the image to a container registry

   ```
   $ make docker-push IMG=quay.io/<username>/gitlab-operator:latest
   ```

 - Run the operator locally to test changes

   ```
   $ make run
   ```

## Deploying the Operator

For instructions on deploying the operator, see the [installation docs](installation.md).
