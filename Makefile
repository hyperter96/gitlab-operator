# Current Operator version
VERSION ?= 0.2.0
# Default bundle image tag
BUNDLE_IMG ?= registry.gitlab.com/gitlab-org/gl-openshift/gitlab-operator-bundle:$(VERSION)
# Options for 'bundle-build'
ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif
ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

# Image URL to use all building/pushing image targets
IMG ?= registry.gitlab.com/gitlab-org/gl-openshift/gitlab-operator
TAG ?= latest
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true,preserveUnknownFields=false"
# Namespace to deploy operator into
NAMESPACE ?= gitlab-system
# Chart version to use in the container
CHART_VERSION ?= $(shell head -n1 CHART_VERSIONS)
# Domain to use for `global.hosts.domain`
DOMAIN ?= example.com
# Host suffix to use for `global.hosts.hostSuffix`
HOSTSUFFIX ?= ""

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

KUSTOMIZE_VERSION ?= 3.8.7

all: manager

# Run tests
test: generate fmt vet manifests
	go test ./... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go

# Install required operators into a cluster
install_required_operators:
	kubectl apply -f scripts/manifests/

# Uninstalls required operators from the cluster
uninstall_required_operators:
	kubectl delete -f scripts/manifests/

# Install CRDs into a cluster
install_crds: manifests kustomize
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall_crds: manifests kustomize
	$(KUSTOMIZE) build config/crd | kubectl delete -f -

# Suffix operator clusterrolebinding names so they can be installed in parallel
suffix_clusterrolebinding_names: kustomize
	cd config/rbac && $(KUSTOMIZE) edit set namesuffix -- "-${NAMESPACE}"

# Suffix operator webhooks names so they can be installed in parallel
suffix_webhook_names: kustomize
	cd config/webhook && $(KUSTOMIZE) edit set namesuffix -- "-${NAMESPACE}"

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy_operator: manifests kustomize
	cd config/manager && $(KUSTOMIZE) edit set image registry.gitlab.com/gitlab-org/gl-openshift/gitlab-operator=${IMG}:${TAG}
	cd config/manager && $(KUSTOMIZE) edit add patch --path patches/deployment_always_pull_image.yaml
	cd config/default && $(KUSTOMIZE) edit set namespace ${NAMESPACE}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

# Delete controller from the configured Kubernetes cluster
delete_operator: manifests kustomize
	cd config/default && $(KUSTOMIZE) edit set namespace ${NAMESPACE}
	$(KUSTOMIZE) build config/default | kubectl delete -f -

# Deploy sample GitLab custom resource to cluster
deploy_sample_cr: kustomize
	cd config/samples && $(KUSTOMIZE) edit set namespace ${NAMESPACE}
	$(KUSTOMIZE) build config/samples \
		| sed "s/CHART_VERSION/${CHART_VERSION}/g" \
		| sed "s/DOMAIN/${DOMAIN}/g" \
		| sed "s/HOSTSUFFIX/${HOSTSUFFIX}/g" \
		| kubectl apply -f -

# Delete the sample GitLab custom resource from cluster
delete_sample_cr: kustomize
	cd config/samples && $(KUSTOMIZE) edit set namespace ${NAMESPACE}
	$(KUSTOMIZE) build config/samples | kubectl delete -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Restores files that are modified during operator and CR deploy
restore_kustomize_files:
	git checkout -q \
    config/default/kustomization.yaml \
    config/manager/kustomization.yaml \
    config/rbac/kustomization.yaml \
    config/samples/kustomization.yaml \
    config/webhook/kustomization.yaml

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Build the docker image
docker-build: test # Pending https://github.com/kubernetes-sigs/kubebuilder/pull/1626
	mkdir -p .go/pkg/mod # for builds outside of CI, this cache directory won't exit
	podman build . -t ${IMG}

# Push the docker image
docker-push:
	podman push ${IMG}

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.3.0 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

kustomize:
ifeq (, $(shell which kustomize))
	@{ \
	set -e ;\
	KUSTOMIZE_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$KUSTOMIZE_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/kustomize/kustomize/v3@v$(KUSTOMIZE_VERSION) ;\
	rm -rf $$KUSTOMIZE_GEN_TMP_DIR ;\
	}
KUSTOMIZE=$(GOBIN)/kustomize
else
KUSTOMIZE=$(shell which kustomize)
endif

# Generate bundle manifests and metadata, then validate generated files.
.PHONY: bundle
bundle: manifests
	operator-sdk generate kustomize manifests -q
	cd config/manager && $(KUSTOMIZE) edit set image ${IMG}=${IMG}:${TAG}
	$(KUSTOMIZE) build config/manifests | operator-sdk generate bundle -q --overwrite --version $(VERSION) $(BUNDLE_METADATA_OPTS)
	operator-sdk bundle validate ./bundle

# Build the bundle image.
.PHONY: bundle-build
bundle-build:
	podman build -f bundle.Dockerfile -t $(BUNDLE_IMG) .

deployment-files: bundle
	cp -av bundle/manifests/apps.gitlab.com_*.yaml config/deploy
	cp -av bundle/manifests/*_serviceaccount.yaml config/deploy
	cp -av bundle/manifests/*_clusterrole.yaml config/deploy
	cp -av bundle/manifests/*_clusterrolebinding.yaml config/deploy
	for rb in `ls config/deploy/*_clusterrolebinding.yaml`; do egrep "  namespace:"  $$rb || echo "  namespace: gitlab-system" >> $$rb; done
	sed -n -e 's/manager-role/gitlab-manager-role/g;w config/deploy/gitlab-manager-role_rbac.authorization.k8s.io_v1_clusterrole.yaml' config/rbac/role.yaml
