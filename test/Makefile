# Include common definitions
include ../Makefile.def

.PHONY: init installDeps deploy installDepsRedfish clean

# Kind cluster configuration
KIND_CLUSTER_NAME ?= bmc-e2e
KIND_CONFIG = kind-config.yaml

# Initialize kind cluster with 1 worker node
init:
	@echo "Creating kind cluster with name: $(KIND_CLUSTER_NAME)"
	kind create cluster --name $(KIND_CLUSTER_NAME) --config $(KIND_CONFIG)
	@echo "Kind cluster created successfully"


installDeps:
	@echo "Installing spiderpool..."
	chmod +x scripts/install-spiderpool.sh
	E2E_CLUSTER_NAME=$(KIND_CLUSTER_NAME) ./scripts/install-spiderpool.sh

installDepsRedfish:
	@echo "Installing redfish..."
	chmod +x scripts/install-redfish.sh
	E2E_CLUSTER_NAME=$(KIND_CLUSTER_NAME) ./scripts/install-redfish.sh


# Deploy application to kind cluster
deploy:
	@echo "Loading images into kind cluster..."
	kind load docker-image $(CONTROLLER_IMAGE_REF) --name $(KIND_CLUSTER_NAME) && \
	kind load docker-image $(AGENT_IMAGE_REF) --name $(KIND_CLUSTER_NAME)
	@echo "Images loaded successfully"
	@echo "Deploying application using Helm chart..."
	- helm uninstall bmc -n bmc &>/dev/null || true
	helm install bmc ../chart \
		--namespace bmc \
		--create-namespace \
		--wait \
		--set image.repository=$(CONTROLLER_IMAGE) \
		--set image.tag=$(VERSION) \
		--set clusterAgent.agentYaml.image.repository=$(AGENT_IMAGE) \
		--set clusterAgent.agentYaml.image.tag=$(VERSION) \
		--set clusterAgent.endpoint.https=false \
		--set clusterAgent.agentYaml.underlayInterface="spiderpool/eth0-macvlan" \
		--set clusterAgent.agentYaml.nodeName="$(KIND_CLUSTER_NAME)-worker" \
		--set clusterAgent.feature.dhcpServerConfig.subnet="192.168.0.0/24" \
		--set clusterAgent.feature.dhcpServerConfig.ipRange="192.168.0.10-192.168.0.100" \
		--set clusterAgent.feature.dhcpServerConfig.gateway="192.168.0.1" \
		--set clusterAgent.feature.dhcpServerConfig.selfIp="192.168.0.2/24" \
		--set clusterAgent.endpoint.username="test" \
		--set clusterAgent.endpoint.password="abc" \
		--set clusterAgent.endpoint.port=8000
		

# Clean up kind cluster
clean:
	@echo "Cleaning up kind cluster: $(KIND_CLUSTER_NAME)"
	@if kind get clusters | grep -q $(KIND_CLUSTER_NAME); then \
		kind delete cluster --name $(KIND_CLUSTER_NAME); \
		echo "Kind cluster deleted successfully"; \
	else \
		echo "Kind cluster $(KIND_CLUSTER_NAME) does not exist"; \
	fi
