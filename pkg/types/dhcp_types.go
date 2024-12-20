package types

// ClientInfo represents DHCP client information
type ClientInfo struct {
	IP             string
	MAC            string
	LeaseStartTime string
	LeaseEndTime   string
	LeaseActive    bool
	HostName       string
}

// IPUsageStats represents DHCP IP usage statistics
type IPUsageStats struct {
	TotalIPs     int
	UsedIPs      int
	UsagePercent float64
}
