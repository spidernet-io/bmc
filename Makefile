# Include common definitions
include Makefile.def

# Version and image registry configuration
VERSION ?= $(shell git rev-parse --short HEAD)
REGISTRY ?= ghcr.io/spidernet-io
IMAGE_NAMESPACE ?= bmc
TOOLS_IMAGE_NAME ?= tools
TOOLS_IMAGE_TAG ?= latest

# Go build configuration
GOOS ?= linux
GOARCH ?= amd64
CGO_ENABLED ?= 0

# Output directory
BIN_DIR := bin

# Default target
.PHONY: all
all: images

# Build targets
.PHONY: build-binaries
build-binaries: build-controller build-agent

.PHONY: build-controller
build-controller:
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BIN_DIR)/controller cmd/controller/main.go

.PHONY: build-agent
build-agent:
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BIN_DIR)/agent cmd/agent/main.go

# Image targets
.PHONY: images
images: controller-image agent-image tools-image

.PHONY: controller-image
controller-image:
	docker build -t $(CONTROLLER_IMAGE_REF) -f image/controller/Dockerfile .

.PHONY: agent-image
agent-image:
	docker build -t $(AGENT_IMAGE_REF) -f image/agent/Dockerfile .

.PHONY: tools-image
tools-image: build-tools-image

.PHONY: build-tools-image
build-tools-image:
	docker build -t $(TOOLS_IMAGE_REF) -f image/tools/Dockerfile image/tools

# Helm chart
.PHONY: chart
chart:
	helm package chart/

# E2E tests
.PHONY: e2e e2e-clean uninstall_e2e

# Run E2E tests
e2e:
	@echo "Setting up E2E test environment..."
	$(MAKE) -C test clean
	$(MAKE) -C test init
	$(MAKE) -C test installDeps
	$(MAKE) -C test deploy
	$(MAKE) -C test installDepsRedfish
	@echo "E2E environment setup completed"

# Clean up E2E environment
e2e-clean:
	@echo "Cleaning up E2E environment..."
	$(MAKE) -C test clean
	@echo "E2E environment cleanup completed"

# Uninstall E2E environment
uninstall_e2e:
	@echo "Uninstalling E2E environment..."
	$(MAKE) -C test kind-delete
	@echo "E2E environment uninstalled successfully"

# Cleanup
.PHONY: clean
clean:
	rm -rf $(BIN_DIR)
	rm -f *.tgz

.PHONY: update_crd_sdk
update_crd_sdk:
	@ echo "update crd manifest" && ./tools/golang/crdControllerGen.sh
	@ echo "update crd sdk" && ./tools/golang/crdSdkGen.sh


.PHONY: validate_crd_sdk
validate_crd_sdk:
	@ echo "validate crd manifest"
	make update_crd_sdk ; \
		if ! test -z "$$(git status --porcelain)"; then \
  			echo "please run 'make update_crd_sdk' to update crd code" ; \
  			exit 1 ; \
  		fi ; echo "succeed to check crd sdk"


# Help
.PHONY: usage
usage:
	@echo "Available targets:"
	@echo "  all             - Build binaries, container images, and Helm chart"
	@echo "  build-binaries  - Build controller and agent binaries"
	@echo "  build-controller - Build controller binary"
	@echo "  build-agent     - Build agent binary"
	@echo "  images          - Build container images"
	@echo "  controller-image - Build controller container image"
	@echo "  agent-image     - Build agent container image"
	@echo "  tools-image     - Build tools container image"
	@echo "  chart           - Package Helm chart"
	@echo "  e2e             - Run E2E tests"
	@echo "  e2e-clean       - Clean up E2E environment"
	@echo "  uninstall_e2e  - Uninstall E2E environment"
	@echo "  clean           - Remove build artifacts"
	@echo "  usage           - Show this help message"
	@echo ""
	@echo "Environment variables:"
	@echo "  VERSION        - Version tag for images (default: from git commit hash)"
	@echo "  REGISTRY       - Container image registry (default: ghcr.io/spidernet-io)"
	@echo "  IMAGE_NAMESPACE - Container image namespace for tools (default: bmc)"
	@echo "  TOOLS_IMAGE_NAME - Name of the tools image (default: tools)"
	@echo "  TOOLS_IMAGE_TAG - Tag of the tools image (default: latest)"
	@echo "  GOOS          - Target OS for build (default: linux)"
	@echo "  GOARCH        - Target architecture for build (default: amd64)"
	@echo "  CGO_ENABLED   - Enable CGO for build (default: 0)"
