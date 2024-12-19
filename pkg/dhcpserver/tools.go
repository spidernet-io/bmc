// Package dhcpserver provides DHCP server management functionality.
package dhcpserver

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/spidernet-io/bmc/pkg/log"
	"github.com/vishvananda/netlink"
)

// configureInterface configures the network interface with the specified IP address
// Parameters:
//   - s.config.DhcpServerInterface: name of the network interface to configure
//   - s.config.SelfIp: optional IP address to assign to the interface
//
// Returns an error if interface configuration fails.
func (s *dhcpServer) configureInterface() error {
	// Get the network interface by name
	link, err := netlink.LinkByName(s.config.DhcpServerInterface)
	if err != nil {
		log.Logger.Errorf("failed to get interface %s: %v", s.config.DhcpServerInterface, err)
		return err
	}

	// Ensure interface is up
	if err := netlink.LinkSetUp(link); err != nil {
		log.Logger.Errorf("failed to set interface %s up: %v", s.config.DhcpServerInterface, err)
		return err
	}

	// Get existing addresses
	addrs, err := netlink.AddrList(link, netlink.FAMILY_V4)
	if err != nil {
		log.Logger.Errorf("failed to get addresses for interface %s: %v", s.config.DhcpServerInterface, err)
		return err
	}

	// Iterate over existing IP addresses and remove them
	for _, addr := range addrs {
		log.Logger.Debugf("Removing IP address %s from interface %s", addr.IP, s.config.DhcpServerInterface)
		if err := netlink.AddrDel(link, &addr); err != nil {
			log.Logger.Errorf("failed to remove IP address %s: %v", addr.IP, err)
			return err
		}
	}

	// Add new IP address if specified
	if s.config.SelfIp != "" {
		addr, err := netlink.ParseAddr(s.config.SelfIp)
		if err != nil {
			log.Logger.Errorf("failed to parse IP address %s: %v", s.config.SelfIp, err)
			return err
		}

		if err := netlink.AddrAdd(link, addr); err != nil {
			log.Logger.Errorf("failed to add IP address %s: %v", s.config.SelfIp, err)
			return err
		}
		log.Logger.Debugf("Added IP address %s to interface %s", s.config.SelfIp, s.config.DhcpServerInterface)
	}
	return nil
}

// generateConfig generates the DHCP server configuration file
// Parameters:
//   - s.config.Subnet: subnet in CIDR notation that the DHCP server will serve
//   - s.config.IpRange: range of IP addresses for allocation
//   - s.config.Gateway: gateway IP address for the subnet
//
// Returns an error if configuration file cannot be created or written.
func (s *dhcpServer) generateConfig() error {
	network, netmask, err := getNetworkAndMask(s.config.Subnet)
	if err != nil {
		return fmt.Errorf("failed to get network and mask: %v", err)
	}

	// Convert IP range format from "start-end" to "start end"
	ipRange := strings.Replace(s.config.IpRange, "-", " ", 1)
	config := fmt.Sprintf(DhcpConfigTemplate,
		network,    // subnet
		netmask,    // netmask in subnet declaration
		ipRange,    // range
		s.config.Gateway,  // routers
		netmask,    // subnet-mask option
	)

	log.Logger.Debugf("Generated DHCP configuration:\n\n%s", config)

	return os.WriteFile(DhcpConfigPath, []byte(config), 0644)
}

// calculateTotalIPs computes the total number of IP addresses available for allocation
// Parameters:
//   - s.config.IpRange: range of IP addresses in format "start end" (e.g., "192.168.1.100 192.168.1.200")
//
// Returns:
//   - s.totalIPs: number of IP addresses in the range
//   - err: error if IP range is invalid or calculation fails
func (s *dhcpServer) calculateTotalIPs() error {
	// Split IP range string into start and end IP addresses
	rangeParts := strings.Split(s.config.IpRange, "-")
	if len(rangeParts) != 2 {
		log.Logger.Errorf("invalid IP range format: %s", s.config.IpRange)
		return fmt.Errorf("invalid IP range format: %s", s.config.IpRange)
	}

	// Parse start and end IP addresses
	startIP := net.ParseIP(strings.TrimSpace(rangeParts[0]))
	endIP := net.ParseIP(strings.TrimSpace(rangeParts[1]))
	if startIP == nil || endIP == nil {
		log.Logger.Errorf("invalid IP addresses in range: %s", s.config.IpRange)
		return fmt.Errorf("invalid IP addresses in range: %s", s.config.IpRange)
	}

	// Calculate total number of IP addresses in the range
	s.totalIPs = ipRange(startIP, endIP)
	log.Logger.Debugf("Calculated total IPs in range %s: %d", s.config.IpRange, s.totalIPs)
	return nil
}

// ipRange calculates the number of IPs in a range
// Parameters:
//   - start: first IP address in the range
//   - end: last IP address in the range
//
// Returns:
//   - total number of IP addresses in the range
func ipRange(start, end net.IP) int {
	// Convert IP addresses to integers
	startInt := ipToInt(start)
	endInt := ipToInt(end)

	// Calculate total number of IP addresses in the range
	return int(endInt - startInt + 1)
}

// ipToInt converts an IP address to an integer
// Parameters:
//   - ip: IP address to convert
//
// Returns:
//   - integer representation of the IP address
func ipToInt(ip net.IP) int64 {
	// Convert IP address to IPv4 format
	ip = ip.To4()
	if ip == nil {
		return 0
	}

	// Convert IP address to integer
	return int64(ip[0])<<24 | int64(ip[1])<<16 | int64(ip[2])<<8 | int64(ip[3])
}

// getNetworkAndMask returns the network and mask from a CIDR notation string
func getNetworkAndMask(cidr string) (string, string, error) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", "", err
	}

	// Convert netmask from IPMask format to dotted decimal format
	mask := fmt.Sprintf("%d.%d.%d.%d",
		ipNet.Mask[0], ipNet.Mask[1], ipNet.Mask[2], ipNet.Mask[3])

	return ipNet.IP.String(), mask, nil
}

// parseLeasesFile parses the DHCP leases file and returns a list of client info
func parseLeasesFile() ([]ClientInfo, error) {
	content, err := os.ReadFile(DhcpLeaseFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	log.Logger.Debugf("Parsing DHCP lease file content:\n%s", string(content))

	var clients []ClientInfo
	lines := strings.Split(string(content), "\n")
	var currentClient *ClientInfo
	var inLeaseBlock bool

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		log.Logger.Debugf("Processing line: %s", line)

		// Start of a lease block
		if strings.HasPrefix(line, "lease") {
			if currentClient != nil {
				log.Logger.Debugf("Adding client: IP=%s, MAC=%s, Active=%v", 
					currentClient.IP, currentClient.MAC, currentClient.Active)
				clients = append(clients, *currentClient)
			}
			currentClient = &ClientInfo{}
			inLeaseBlock = true
			
			// Extract IP address
			parts := strings.Split(line, " ")
			if len(parts) >= 2 {
				currentClient.IP = strings.TrimSpace(strings.TrimSuffix(parts[1], "{"))
				log.Logger.Debugf("Found lease for IP: %s", currentClient.IP)
			}
			continue
		}

		// End of a lease block
		if line == "}" {
			if currentClient != nil {
				clients = append(clients, *currentClient)
				currentClient = nil
			}
			inLeaseBlock = false
			continue
		}

		// Inside a lease block
		if inLeaseBlock && currentClient != nil {
			// Extract MAC address
			if strings.Contains(line, "hardware ethernet") {
				parts := strings.Split(line, "hardware ethernet")
				if len(parts) == 2 {
					currentClient.MAC = strings.TrimSpace(strings.TrimSuffix(parts[1], ";"))
				}
			}

			// Extract binding state
			if strings.Contains(line, "binding state") {
				parts := strings.Split(line, "binding state")
				if len(parts) == 2 {
					state := strings.TrimSpace(strings.TrimSuffix(parts[1], ";"))
					currentClient.Active = (state == "active")
				}
			}

			// Extract start time
			if strings.Contains(line, "starts") {
				parts := strings.Fields(line)
				if len(parts) >= 4 {
					currentClient.StartTime = strings.TrimSuffix(parts[2] + " " + parts[3], ";")
				}
			}

			// Extract end time
			if strings.Contains(line, "ends") {
				parts := strings.Fields(line)
				if len(parts) >= 4 {
					currentClient.EndTime = strings.TrimSuffix(parts[2] + " " + parts[3], ";")
				}
			}
		}
	}

	// Handle the last lease block if exists
	if currentClient != nil {
		log.Logger.Debugf("Adding final client: IP=%s, MAC=%s, Active=%v",
			currentClient.IP, currentClient.MAC, currentClient.Active)
		clients = append(clients, *currentClient)
	}

	log.Logger.Debugf("Total clients found: %d", len(clients))
	return clients, nil
}

// parseDhcpTime parses time from DHCP lease file format
// Format example: "starts 3 2024/12/18 10:00:00;"
func parseDhcpTime(line string) (time.Time, error) {
	parts := strings.Fields(line)
	if len(parts) < 4 {
		return time.Time{}, fmt.Errorf("invalid time format")
	}

	// Combine date and time parts
	timeStr := parts[2] + " " + parts[3]
	// Remove trailing semicolon if present
	timeStr = strings.TrimSuffix(timeStr, ";")

	// Parse the combined string
	return time.Parse("2006/01/02 15:04:05", timeStr)
}
