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
- role: worker
endef
export KIND_CONFIG_CONTENT

# Initialize kind cluster with 2 worker nodes
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
	@IMAGE=$$(grep "repository:" ../chart/values.yaml | awk '{print $$2}'); \
	TAG=$$(grep "tag:" ../chart/values.yaml | awk '{print $$2}' | sed 's/"//g'); \
	if [ -z "$$TAG" ]; then \
		TAG=$$(grep "appVersion:" ../chart/Chart.yaml | awk '{print $$2}' | sed 's/"//g'); \
	fi; \
	FULL_IMAGE="$$IMAGE:$$TAG"; \
	echo "Loading image: $$FULL_IMAGE"; \
	kind load docker-image $$FULL_IMAGE --name $(KIND_CLUSTER_NAME)
	@echo "Images loaded successfully"
	@echo "Deploying application using Helm chart..."
	helm upgrade --install bmc ../chart --namespace bmc --create-namespace

# Clean up kind cluster
clean:
	@echo "Cleaning up kind cluster: $(KIND_CLUSTER_NAME)"
	@if kind get clusters | grep -q $(KIND_CLUSTER_NAME); then \
		kind delete cluster --name $(KIND_CLUSTER_NAME); \
		echo "Kind cluster deleted successfully"; \
	else \
		echo "Kind cluster $(KIND_CLUSTER_NAME) does not exist"; \
	fi
