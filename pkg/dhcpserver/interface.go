// Package dhcpserver provides DHCP server management functionality.
package dhcpserver

// ClientInfo represents information about a DHCP client lease
type ClientInfo struct {
	// IP is the allocated IP address for the client
	IP string
	// MAC is the client's MAC address
	MAC string
	// Active indicates whether this lease is in active state
	Active bool
	// StartTime is when the lease starts (in format "2024/12/18 10:00:00")
	StartTime string
	// EndTime is when the lease ends (in format "2024/12/18 10:30:00")
	EndTime string
}

// IPUsageStats represents statistics about IP address allocation
type IPUsageStats struct {
	// TotalIPs is the total number of IP addresses available for allocation
	TotalIPs int
	// AvailableIPs is the number of IP addresses currently available
	AvailableIPs int
}

// DhcpServer defines the interface for DHCP server operations.
// This interface provides methods to control the DHCP server and retrieve information
// about its current state.
type DhcpServer interface {
	// Start initializes and starts the DHCP server.
	// It configures the network interface if SelfIp is specified,
	// generates the DHCP configuration, and starts the dhcpd process.
	// Returns an error if any step fails.
	Start() error

	// Stop gracefully stops the DHCP server.
	// It terminates the dhcpd process and cleans up any monitoring routines.
	// Returns an error if the server cannot be stopped.
	Stop() error

	// GetClientInfo retrieves information about all current DHCP clients.
	// Returns a list of ClientInfo containing IP and MAC addresses for each client,
	// or an error if the information cannot be retrieved.
	GetClientInfo() ([]ClientInfo, error)

	// GetIPUsageStats retrieves current IP allocation statistics.
	// Returns IPUsageStats containing total and available IP counts,
	// or an error if the statistics cannot be calculated.
	GetIPUsageStats() (*IPUsageStats, error)
}
