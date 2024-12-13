# BMC Kubernetes Operator

A Kubernetes operator for managing BMC (Baseboard Management Controller) servers in a Kubernetes cluster.

## Features

- Custom Resource Definition (CRD) for managing BMC servers
- Automatic deployment of BMC server pods with network interface configuration
- Helm-based installation and configuration
- Support for multiple BMC server instances
- Network interface configuration via CNI annotations
- Cluster name configuration via environment variables

## Prerequisites

- Kubernetes cluster (v1.16+)
- Helm v3
- kubectl configured to communicate with your cluster

## Installation

1. Add the Helm repository:
```bash
helm repo add bmc https://[your-helm-repo-url]
helm repo update
```

2. Install the operator:
```bash
helm install bmc-operator bmc/bmc-operator
```

### Configuration

You can customize the installation by creating a `values.yaml` file:

```yaml
operator:
  replicas: 2
  image:
    repository: bmc/operator
    tag: latest

server:
  image:
    repository: bmc/server
    tag: latest
```

Then install with:
```bash
helm install -f values.yaml bmc-operator bmc/bmc-operator
```

## Usage

1. Create a BMC Server instance:

```yaml
apiVersion: bmc.io/v1beta1
kind: bmcServer
metadata:
  name: test-bmc
spec:
  interface: "kube-system/macvlan-pod-network"
  clusterName: "cluster1"
```

2. Apply the configuration:
```bash
kubectl apply -f bmc-server.yaml
```

## Development

### Building from Source

1. Prerequisites:
   - Go 1.19+
   - Docker
   - Make

2. Build all components:
```bash
make all
```

3. Build specific components:
```bash
make operator-image  # Build operator image
make server-image   # Build server image
make chart          # Package Helm chart
```

For more build options:
```bash
make usage
```

## License

[Your License]
