.PHONY: init deploy clean

# Kind cluster configuration
KIND_CLUSTER_NAME ?= bmc-e2e
KIND_CONFIG = kind-config.yaml

# Create kind cluster configuration file
define KIND_CONFIG_CONTENT
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
- role: worker
endef
export KIND_CONFIG_CONTENT

# Initialize kind cluster with 1 worker node
init:
	@echo "Creating kind cluster configuration..."
	@echo "$$KIND_CONFIG_CONTENT" > $(KIND_CONFIG)
	@echo "Creating kind cluster with name: $(KIND_CLUSTER_NAME)"
	kind create cluster --name $(KIND_CLUSTER_NAME) --config $(KIND_CONFIG)
	@rm -f $(KIND_CONFIG)
	@echo "Kind cluster created successfully"

# Deploy application to kind cluster
deploy:
	@echo "Loading images into kind cluster..."
	@GIT_HASH=$$(git rev-parse --short HEAD); \
	kind load docker-image spidernet-io/bmc/controller:$$GIT_HASH --name $(KIND_CLUSTER_NAME) && \
	kind load docker-image spidernet-io/bmc/agent:$$GIT_HASH --name $(KIND_CLUSTER_NAME)
	@echo "Images loaded successfully"
	@echo "Deploying application using Helm chart..."
	@GIT_HASH=$$(git rev-parse --short HEAD); \
	helm upgrade --install bmc ../chart --namespace bmc --create-namespace \
		--set image.tag=$$GIT_HASH \
		--set clusterAgent.image.tag=$$GIT_HASH

# Clean up kind cluster
clean:
	@echo "Cleaning up kind cluster: $(KIND_CLUSTER_NAME)"
	@if kind get clusters | grep -q $(KIND_CLUSTER_NAME); then \
		kind delete cluster --name $(KIND_CLUSTER_NAME); \
		echo "Kind cluster deleted successfully"; \
	else \
		echo "Kind cluster $(KIND_CLUSTER_NAME) does not exist"; \
	fi
