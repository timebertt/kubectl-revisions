TOOLS_BIN_DIR ?= hack/tools/bin
export PATH := $(abspath $(TOOLS_BIN_DIR)):$(PATH)

# We use a file per tool and version as an indicator for make whether we need to install the tool or a different version
# of the tool (make doesn't rerun the rule if the rule is changed).

# Use this "function" to add the version file as a prerequisite for the tool target, e.g.:
#   $(KUBECTL): $(call tool_version_file,$(KUBECTL),$(KUBECTL_VERSION))
tool_version_file = $(TOOLS_BIN_DIR)/.version_$(subst $(TOOLS_BIN_DIR)/,,$(1))_$(2)

# Use this "function" to get the version of a go module from go.mod, e.g.:
#   GINKGO_VERSION ?= $(call version_gomod,github.com/onsi/ginkgo/v2)
version_gomod = $(shell go list -f '{{ .Version }}' -m $(1))

# This target cleans up any previous version files for the given tool and creates the given version file.
# This way, we can generically determine, which version was installed without calling each and every binary explicitly.
$(TOOLS_BIN_DIR)/.version_%:
	@version_file=$@; rm -f $${version_file%_*}*
	@touch $@

DYFF := $(TOOLS_BIN_DIR)/dyff
# renovate: datasource=github-releases depName=homeport/dyff
DYFF_VERSION ?= 1.10.2
$(DYFF): $(call tool_version_file,$(DYFF),$(DYFF_VERSION))
	curl -sSfL https://github.com/homeport/dyff/releases/download/v$(DYFF_VERSION)/dyff_$(DYFF_VERSION)_$(shell uname -s | tr '[:upper:]' '[:lower:]')_$(shell uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/').tar.gz | tar -xzmf - -C $(TOOLS_BIN_DIR) dyff

GINKGO := $(TOOLS_BIN_DIR)/ginkgo
GINKGO_VERSION ?= $(call version_gomod,github.com/onsi/ginkgo/v2)
$(GINKGO): $(call tool_version_file,$(GINKGO),$(GINKGO_VERSION))
	go build -o $(GINKGO) github.com/onsi/ginkgo/v2/ginkgo

GOIMPORTS_REVISER := $(TOOLS_BIN_DIR)/goimports-reviser
# renovate: datasource=github-releases depName=incu6us/goimports-reviser
GOIMPORTS_REVISER_VERSION ?= v3.10.0
$(GOIMPORTS_REVISER): $(call tool_version_file,$(GOIMPORTS_REVISER),$(GOIMPORTS_REVISER_VERSION))
	GOBIN=$(abspath $(TOOLS_BIN_DIR)) go install github.com/incu6us/goimports-reviser/v3@$(GOIMPORTS_REVISER_VERSION)

GOLANGCI_LINT := $(TOOLS_BIN_DIR)/golangci-lint
# renovate: datasource=github-releases depName=golangci/golangci-lint
GOLANGCI_LINT_VERSION ?= v2.4.0
$(GOLANGCI_LINT): $(call tool_version_file,$(GOLANGCI_LINT),$(GOLANGCI_LINT_VERSION))
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(TOOLS_BIN_DIR) $(GOLANGCI_LINT_VERSION)

GORELEASER := $(TOOLS_BIN_DIR)/goreleaser
# renovate: datasource=github-releases depName=goreleaser/goreleaser
GORELEASER_VERSION ?= v2.12.0
$(GORELEASER): $(call tool_version_file,$(GORELEASER),$(GORELEASER_VERSION))
	curl -sSfL https://github.com/goreleaser/goreleaser/releases/download/$(GORELEASER_VERSION)/goreleaser_$(shell uname -s)_$(shell uname -m).tar.gz | tar -xzmf - -C $(TOOLS_BIN_DIR) goreleaser

KIND := $(TOOLS_BIN_DIR)/kind
# renovate: datasource=github-releases depName=kubernetes-sigs/kind
KIND_VERSION ?= v0.30.0
$(KIND): $(call tool_version_file,$(KIND),$(KIND_VERSION))
	curl -Lo $(KIND) https://kind.sigs.k8s.io/dl/$(KIND_VERSION)/kind-$(shell uname -s | tr '[:upper:]' '[:lower:]')-$(shell uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
	chmod +x $(KIND)

KUBECTL := $(TOOLS_BIN_DIR)/kubectl
# renovate: datasource=github-releases depName=kubectl packageName=kubernetes/kubernetes
KUBECTL_VERSION ?= v1.34.1
$(KUBECTL): $(call tool_version_file,$(KUBECTL),$(KUBECTL_VERSION))
	curl -Lo $(KUBECTL) https://dl.k8s.io/release/$(KUBECTL_VERSION)/bin/$(shell uname -s | tr '[:upper:]' '[:lower:]')/$(shell uname -m | sed 's/x86_64/amd64/')/kubectl
	chmod +x $(KUBECTL)
