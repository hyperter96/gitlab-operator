# Developer Guide

This developer guide aims to walk a new developer on how to setup up their environment to be able to contribute to this project.

## Setting up development environment

To setup your system for development of the operator, follow the steps below:

1. Install [golang](https://go.dev/dl/) in your environment.

   ```shell
   go version
   ```

1. Download `operator-sdk`. Current `operator-sdk` [releases](https://github.com/operator-framework/operator-sdk/releases) can be found in the projects repository.

   To check your version of operator SDK run,

   ```shell
   operator-sdk version
   ```

   To contribute code to the current GitLab Operator release, you will need at least operator SDK v1.0.0.

1. Install `task` per the [official installation instructions](https://taskfile.dev/#/installation).
   We use [`task` in place of `make`](https://gitlab.com/gitlab-org/cloud-native/gitlab-operator/-/blob/master/doc/adr/0016-replace-makefile-with-taskfile.md)
   for this project.

1. Clone the `gitlab-operator` repository into your GOPATH.

   ```shell
   git clone git@gitlab.com:gitlab-org/cloud-native/gitlab-operator.git
   ```

## Project structure

The GitLab Operator is built using the Operator SDK v1.0.0 and consequently uses the Kubebuilder v2 layout format. This is necessary to know since there was a change in project directory and some of the tooling used by operator SDK.

```shell
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

- The `controllers` directory contains the controller implementations for the GitLab and GitLab Backup controllers.
- The `api` directory contains the API resource definitions for the GitLab and GLBackup resources owned by the operator. The API definitions are grouped by their API version.
  The `*_types.go` file inside `api/<api_version>` contains spec definitions and markers used to generate the Custom Resource Definitions and Cluster Service Version file used by OLM.
- The `config/samples` directory contains an example manifest for the GitLab Custom Resource.
- The `config/test` directory contains a parametrized GitLab definition used for running integration tests.

  An example is shown below:

  `// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete`

  The contents of `config/rbac/custom` were created manually and is not affected by the RBAC markers.

  Most of the other contents of the config directory are automatically generated but could be modified using `kustomize`.

- The `hack/assets` path contains resources that would need to be pushed inside the operator image when the container image is being built. This is where release files would go.

### Additional Resources

The `Taskfile` allows us to customize manage different tasks such as:

- Creating an Operator Lifecycle Manager bundle

  ```shell
  task bundle
  ```

- Building a container image for the operator

  ```shell
  task docker-build IMG=quay.io/<username>/gitlab-operator:latest
  ```

- Pushing the image to a container registry

  ```shell
  task docker-push IMG=quay.io/<username>/gitlab-operator:latest
  ```

- Run the operator locally to test changes

  ```shell
  task run
  ```

- Run unit tests locally in Docker:

  ```shell
  make test-in-docker
  ```

- Run unit tests locally in Docker, skipping the slow controller tests:

  ```shell
  make unit-tests-in-docker
  ```

- Run unit tests locally in Docker, focusing on the slow controller tests:

  ```shell
  make slow-unit-tests-in-docker
  ```

- Clean up artifacts from local tests in Docker:

  ```shell
  make test-docker-clean
  ```

## Deploying the Operator

For instructions on deploying the operator, see the [installation docs](installation.md).

## Debugging

There have been a couple of functions added to `controllers/gitlab/template_test.go`
to assist in the development of features and the writing of tests.

- `dumpTemplate(template)`
- `dumpTemplateToFile(template, filename)`
- `dumpHelmValues(values)`
- `dumpHelmValuesToFile(values, filename)`

The `dumpTemplate()` function will take the template object from the GitLab
adapter and return the rendered YAML of the Helm chart as a string. Since
the Go test framework will absorb anything written to stdout, the
`dumpTemplateToFile()` will write the YAML to a file for inspection. It
is important to note that if just a filename is provided that the file will
be written to the subdirectory where the test file resides rather than the
directory where the tests were initiated from. An absolute file path is
necessary if one desires the file to be written where the tests are
initiated from.

Similarly the `dumpHelmValues()` will return the YAML representation of the
Helm values as string. This is can be used to verify that the intended
values are set at the beginning of any tests. The `dumpHelmValues()` function
is used to write the YAML to a file for inspection and the filename argument
has the same limitations as `dumpTemplateToFile()`.
