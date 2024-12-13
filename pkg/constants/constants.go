package constants

const (
	// Finalizer name for ClusterAgent
	ClusterAgentFinalizer = "bmc.spidernet.io/finalizer"

	// Label keys
	LabelApp         = "app"
	LabelController  = "controller"
	LabelClusterName = "clusterName"

	// Label values
	LabelValueBMCAgent = "bmc-agent"

	// Environment variables
	EnvPodNamespace = "POD_NAMESPACE"
	EnvAgentImage   = "agentImage"
	EnvClusterName  = "ClusterName"

	// Resource names
	AgentNamePrefix = "agent-"

	// Port names and numbers
	PortHealth     = "health"
	PortNumber     = 8000
	PortProtocol   = "TCP"

	// Health check paths
	HealthCheckPath = "/healthz"

	// Network annotation
	NetworkAnnotationKey = "k8s.v1.cni.cncf.io/networks"
)
