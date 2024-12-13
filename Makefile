# Version and image registry configuration
VERSION := $(shell cat VERSION)
REGISTRY ?= spidernet-io/bmc

# Go build configuration
GOOS ?= linux
GOARCH ?= amd64
CGO_ENABLED ?= 0

# Output directory
BIN_DIR := bin

# Default target
.PHONY: all
all: build images chart

# Build targets
.PHONY: build
build: build-controller build-agent

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
	docker build -t $(REGISTRY)/server:$(VERSION) -f image/agent/Dockerfile .

# Helm chart
.PHONY: chart
chart:
	helm package chart/

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
	@echo "  build           - Build controller and agent binaries"
	@echo "  build-controller - Build controller binary"
	@echo "  build-agent     - Build agent binary"
	@echo "  images          - Build container images"
	@echo "  controller-image - Build controller container image"
	@echo "  agent-image     - Build agent container image"
	@echo "  chart           - Package Helm chart"
	@echo "  clean           - Remove build artifacts"
	@echo "  usage           - Show this help message"
	@echo ""
	@echo "Environment variables:"
	@echo "  VERSION        - Version tag for images (default: from VERSION file)"
	@echo "  REGISTRY       - Container image registry (default: spidernet-io/bmc)"
	@echo "  GOOS          - Target OS for build (default: linux)"
	@echo "  GOARCH        - Target architecture for build (default: amd64)"
	@echo "  CGO_ENABLED   - Enable CGO for build (default: 0)"
