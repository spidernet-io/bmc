# Version and image registry configuration
VERSION ?= $(shell git rev-parse --short HEAD)
REGISTRY ?= spidernet-io/bmc

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
images: controller-image agent-image

.PHONY: controller-image
controller-image:
	docker build -t $(REGISTRY)/controller:$(VERSION) -f image/controller/Dockerfile .

.PHONY: agent-image
agent-image:
	docker build -t $(REGISTRY)/agent:$(VERSION) -f image/agent/Dockerfile .

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
	$(MAKE) -C test deploy
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
	@echo "  chart           - Package Helm chart"
	@echo "  e2e             - Run E2E tests"
	@echo "  e2e-clean       - Clean up E2E environment"
	@echo "  uninstall_e2e  - Uninstall E2E environment"
	@echo "  clean           - Remove build artifacts"
	@echo "  usage           - Show this help message"
	@echo ""
	@echo "Environment variables:"
	@echo "  VERSION        - Version tag for images (default: from git commit hash)"
	@echo "  REGISTRY       - Container image registry (default: spidernet-io/bmc)"
	@echo "  GOOS          - Target OS for build (default: linux)"
	@echo "  GOARCH        - Target architecture for build (default: amd64)"
	@echo "  CGO_ENABLED   - Enable CGO for build (default: 0)"
