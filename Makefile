KUSTOMIZE ?= kustomize
KUBECTL ?= kubectl
# Current Operator version
VERSION ?= 0.2.0
# Default bundle image tag
BUNDLE_IMG ?= registry.gitlab.com/gitlab-org/cloud-native/gitlab-operator-bundle:$(VERSION)
# Options for 'bundle-build'
ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif
ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)
BUNDLE_OPTS ?= --extra-service-accounts=gitlab-manager,gitlab-nginx-ingress,gitlab-app

BUILD_DIR ?= .build
INSTALL_DIR ?= .install

KUSTOMIZE_FILES=$(shell find config -type f -name \*.yaml)
TEST_CR_FILES=$(shell find config/test -type f -name \*.yaml)

# Image URL to use all building/pushing image targets
DEFAULT_IMG := registry.gitlab.com/gitlab-org/cloud-native/gitlab-operator
IMG ?= registry.gitlab.com/gitlab-org/cloud-native/gitlab-operator
TAG ?= latest

# Namespace to deploy operator into
NAMESPACE ?= gitlab-system
# Chart version to use in the container
CHART_VERSION ?= $(shell head -n1 CHART_VERSIONS)
# Domain to use for `global.hosts.domain`
DOMAIN ?= example.com
# Host suffix to use for `global.hosts.hostSuffix`
HOSTSUFFIX ?= ""
# TLS secret name to use for `global.ingress.tls.secretName`
TLSSECRETNAME ?= ""

# Platform for operator deployment, kubernetes or openshift
PLATFORM ?= kubernetes

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

$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

# Build CRDs
$(BUILD_DIR)/crds.yaml: $(BUILD_DIR) manifests kustomize $(KUSTOMIZE_FILES)
	$(KUSTOMIZE) build config/crd > $@

build_crds: $(BUILD_DIR)/crds.yaml

$(INSTALL_DIR):
	mkdir -p $(INSTALL_DIR)

$(INSTALL_DIR)/crds.yaml: $(BUILD_DIR)/crds.yaml $(INSTALL_DIR)
	$(KUBECTL) apply -f $<
	cp $< $@

# Install CRDs into a cluster
install_crds: $(INSTALL_DIR)/crds.yaml

# Uninstall CRDs from a cluster
uninstall_crds: manifests kustomize build_crds
	$(KUBECTL) delete -f $(BUILD_DIR)/crds.yaml
	rm $(INSTALL_DIR)/crds.yaml

# Suffix operator clusterrolebinding names so they can be installed in parallel
suffix_clusterrolebinding_names: kustomize
	cd config/rbac && $(KUSTOMIZE) edit set namesuffix -- "-${NAMESPACE}"

# Suffix operator webhooks names so they can be installed in parallel
suffix_webhook_names: kustomize
	cd config/webhook && $(KUSTOMIZE) edit set namesuffix -- "-${NAMESPACE}"

$(BUILD_DIR)/openshift_resources.yaml: $(BUILD_DIR) $(KUSTOMIZE_FILES)
	cd config/openshift && $(KUSTOMIZE) edit set namespace ${NAMESPACE}
	$(KUSTOMIZE) build config/openshift > $@

build_openshift_resources: $(BUILD_DIR)/openshift_resources.yaml

$(INSTALL_DIR)/openshift_resources.yaml: $(BUILD_DIR)/openshift_resources.yaml $(INSTALL_DIR)
	$(KUBECTL) create namespace ${NAMESPACE} || true
	$(KUBECTL) label namespace ${NAMESPACE} control-plane=controller-manager || true
	$(KUBECTL) apply -f $<
	cp $< $@

deploy_openshift_resources: $(INSTALL_DIR)/openshift_resources.yaml

$(BUILD_DIR)/operator.yaml: $(BUILD_DIR) $(KUSTOMIZE_FILES)
	cd config/default && $(KUSTOMIZE) edit set namespace ${NAMESPACE}
	cd config/manager && $(KUSTOMIZE) edit set image $(DEFAULT_IMG)=$(IMG):$(TAG)
	$(KUSTOMIZE) build config/default > $@

build_operator: $(BUILD_DIR)/operator.yaml

${BUILD_DIR}/yaml-separator:
	echo "---" > $@

${BUILD_DIR}/operator-openshift.yaml: ${BUILD_DIR}/operator.yaml ${BUILD_DIR}/openshift_resources.yaml ${BUILD_DIR}/yaml-separator
	cat ${BUILD_DIR}/operator.yaml ${BUILD_DIR}/yaml-separator ${BUILD_DIR}/openshift_resources.yaml > $@

build_operator_openshift: $(BUILD_DIR)/operator-openshift.yaml

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
$(INSTALL_DIR)/operator.yaml: $(BUILD_DIR)/operator.yaml $(INSTALL_DIR)
	$(KUBECTL) create namespace ${NAMESPACE} || true
	$(KUBECTL) label namespace ${NAMESPACE} control-plane=controller-manager || true
	$(KUBECTL) apply -f $<
	cp $< $@

deploy_operator: $(INSTALL_DIR)/operator.yaml

# Delete controller from the configured Kubernetes cluster
delete_operator: manifests kustomize $(BUILD_DIR)/operator.yaml
	$(KUBECTL) delete -f $(BUILD_DIR)/operator.yaml
	rm $(INSTALL_DIR)/operator.yaml

# Deploy test GitLab custom resource to cluster
build_test_cr: $(BUILD_DIR)/test_cr.yaml

$(BUILD_DIR)/test_cr.yaml: $(BUILD_DIR) $(TEST_CR_FILES)
	cd config/test && $(KUSTOMIZE) edit set namespace ${NAMESPACE}
	$(KUSTOMIZE) build config/test \
		| sed "s/CHART_VERSION/${CHART_VERSION}/g" \
		| sed "s/DOMAIN/${DOMAIN}/g" \
		| sed "s/HOSTSUFFIX/${HOSTSUFFIX}/g" \
		| sed "s/TLSSECRETNAME/${TLSSECRETNAME}/g" > $@

# Deploy test GitLab custom resource to cluster
$(INSTALL_DIR)/test_cr.yaml: $(BUILD_DIR)/test_cr.yaml
	kubectl apply -f $<
	cp $< $@

deploy_test_cr: $(INSTALL_DIR)/test_cr.yaml

# Delete the test GitLab custom resource from cluster
delete_test_cr: $(INSTALL_DIR)/test_cr.yaml
	kubectl delete -f $<

# Restores files that are modified during operator and CR deploy
restore_kustomize_files:
	git checkout -q \
    config/default/kustomization.yaml \
    config/manager/kustomization.yaml \
    config/openshift/kustomization.yaml \
    config/rbac/kustomization.yaml \
    config/test/kustomization.yaml \
    config/webhook/kustomization.yaml

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code

.PHONY: manifests
manifests:
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: generate
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Build the docker image
docker-build: test # Pending https://github.com/kubernetes-sigs/kubebuilder/pull/1626
	mkdir -p .go/pkg/mod # for builds outside of CI, this cache directory won't exit
	podman build . -t $(IMG):$(TAG)

# Push the docker image
docker-push:
	podman push $(IMG):$(TAG)

CONTROLLER_GEN = $(shell which controller-gen)
.PHONY: controller-gen
controller-gen: ## Download controller-gen locally if necessary.
    $(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.7.0)

KUSTOMIZE = $(shell which kustomize)
.PHONY: kustomize
kustomize:
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v3@v3.8.7)

ENVTEST = $(shell which setup-envtest)
.PHONY: envtest
envtest: ## Download envtest-setup locally if necessary.
	$(call go-get-tool,$(ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest@latest)

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR) $(INSTALL_DIR)

# Generate bundle manifests and metadata, then validate generated files.
.PHONY: bundle
bundle: manifests
	operator-sdk generate kustomize manifests -q
	cd config/manager && $(KUSTOMIZE) edit set image $(IMG)=$(IMG):$(TAG)
	$(KUSTOMIZE) build config/manifests | operator-sdk generate bundle -q --overwrite --version $(VERSION) $(BUNDLE_OPS) $(BUNDLE_METADATA_OPTS)
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

define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
go get $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef
