PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Tools

include hack/tools.mk

.PHONY: clean-tools-bin
clean-tools-bin: ## Empty the tools binary directory.
	rm -rf $(TOOLS_BIN_DIR)/*

##@ Development

.PHONY: fmt
fmt: $(GOIMPORTS_REVISER) ## Run go fmt against code.
	go fmt ./...
	$(GOIMPORTS_REVISER) ./...

.PHONY: modules
modules: ## Runs go mod to ensure modules are up to date.
	go mod tidy

.PHONY: test
test: ## Run tests.
	go test -race ./pkg/...

.PHONY: test-cover
test-cover: ## Run tests with coverage.
	go test -coverprofile cover.out ./...
	go tool cover -html cover.out -o cover.html

GINKGO_FLAGS ?=
TEST_FLAGS ?=

.PHONY: test-e2e
test-e2e: $(GINKGO) $(KUBECTL) ## Run e2e tests.
	ginkgo run --timeout=10m --poll-progress-after=10s --poll-progress-interval=5s --randomize-all --randomize-suites --keep-going --show-node-events $(GINKGO_FLAGS) ./test/e2e/... -- $(TEST_FLAGS)

##@ Verification

.PHONY: lint
lint: $(GOLANGCI_LINT) ## Run golangci-lint against code.
	$(GOLANGCI_LINT) run ./...

.PHONY: check
check: lint test ## Check everything (lint + test).

.PHONY: verify-fmt
verify-fmt: fmt ## Verify go code is formatted.
	@if !(git diff --quiet HEAD); then \
		echo "unformatted files detected, please run 'make fmt'"; exit 1; \
	fi

.PHONY: verify-modules
verify-modules: modules ## Verify go module files are up to date.
	@if !(git diff --quiet HEAD -- go.sum go.mod); then \
		echo "go module files are out of date, please run 'make modules'"; exit 1; \
	fi

.PHONY: verify-goreleaser
verify-goreleaser: $(GORELEASER) ## Verify .goreleaser.yaml
	$(GORELEASER) check

.PHONY: verify
verify: verify-fmt verify-modules verify-goreleaser check ## Verify everything (all verify-* rules + check).

.PHONY: ci-e2e-kind
ci-e2e-kind: $(KIND)
	./hack/ci-e2e-kind.sh

##@ Build

.PHONY: build
build: ## Build the kubectl-history binary.
	go build -o bin/kubectl-history .

.PHONY: install
install: ## Install the kubectl-history binary to $GOBIN.
	go install .

##@ Test Setup

# renovate: datasource=docker depName=kindest/node
KIND_KUBERNETES_VERSION ?= v1.28.9

KIND_KUBECONFIG := $(PROJECT_DIR)/hack/kind_kubeconfig.yaml
kind-up kind-down: export KUBECONFIG = $(KIND_KUBECONFIG)

.PHONY: kind-up
kind-up: $(KIND) $(KUBECTL) ## Launch a kind cluster for local development and testing.
	$(KIND) create cluster --name history --image kindest/node:$(KIND_KUBERNETES_VERSION)
	# workaround https://kind.sigs.k8s.io/docs/user/known-issues/#pod-errors-due-to-too-many-open-files
	$(KUBECTL) get nodes -o name | cut -d/ -f2 | xargs -I {} docker exec {} sh -c "sysctl fs.inotify.max_user_instances=8192"
	# run `export KUBECONFIG=$$PWD/hack/kind_kubeconfig.yaml` to target the created kind cluster.

.PHONY: kind-down
kind-down: $(KIND) ## Tear down the kind testing cluster.
	$(KIND) delete cluster --name history
