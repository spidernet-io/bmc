// Package dhcpserver provides DHCP server management functionality.
package dhcpserver

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"time"

	"github.com/spidernet-io/bmc/pkg/dhcpserver/types"
	bmcv1beta1 "github.com/spidernet-io/bmc/pkg/k8s/apis/bmc.spidernet.io/v1beta1"
	"github.com/spidernet-io/bmc/pkg/lock"
	"github.com/spidernet-io/bmc/pkg/log"
)

// DhcpServer defines the interface for DHCP server operations
type DhcpServer interface {
	Start() error
	Stop() error
	GetClientInfo() ([]types.ClientInfo, error)
	GetIPUsageStats() (*types.IPUsageStats, error)
}

// dhcpServer implements the DhcpServer interface.
// It manages the lifecycle of an ISC DHCP server instance and provides
// methods to monitor its operation and retrieve statistics.
type dhcpServer struct {
	// config holds the DHCP server configuration
	config *bmcv1beta1.DhcpServerConfig
	// cmd represents the running DHCP server process
	cmd *exec.Cmd
	// mutex protects access to shared resources
	mutex lock.Mutex
	// stopChan signals the monitoring routine to stop
	stopChan chan struct{}
	// stats holds current IP usage statistics
	stats types.IPUsageStats
	// totalIPs is the total number of IP addresses available for allocation
	totalIPs int
	// ipRangeErr stores any error encountered while parsing IP range
	ipRangeErr error
	// previousClients stores the last known state of DHCP clients
	previousClients []types.ClientInfo
	// clusterAgentName is used to generate unique lease file path
	clusterAgentName string
	// leaseFilePath is the rendered lease file path
	leaseFilePath string
	// Event channels for hoststatus
	addChan    chan<- types.ClientInfo
	deleteChan chan<- types.ClientInfo
}

var _ DhcpServer = (*dhcpServer)(nil)

// NewDhcpServer creates a new DHCP server instance.
// Parameters:
//   - config: DHCP server configuration including interface, subnet, and IP range
//   - clusterAgentName: name used to generate unique lease file path
//   - addChan: channel for notifying about new or modified DHCP clients
//   - deleteChan: channel for notifying about removed DHCP clients
//
// Returns:
//   - DhcpServer interface implementation
func NewDhcpServer(config *bmcv1beta1.DhcpServerConfig, clusterAgentName string, addChan chan<- types.ClientInfo, deleteChan chan<- types.ClientInfo) (*dhcpServer, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if clusterAgentName == "" {
		return nil, fmt.Errorf("clusterAgentName cannot be empty")
	}

	// Check if interface exists and is up
	iface, err := net.InterfaceByName(config.DhcpServerInterface)
	if err != nil {
		return nil, fmt.Errorf("interface %s not found: %v", config.DhcpServerInterface, err)
	}

	// Check if interface is up
	if iface.Flags&net.FlagUp == 0 {
		return nil, fmt.Errorf("interface %s is down", config.DhcpServerInterface)
	}

	if config.SelfIp != "" {
		// Check if SelfIP is within the subnet
		selfIP := net.ParseIP(strings.Split(config.SelfIp, "/")[0])
		if selfIP == nil {
			return nil, fmt.Errorf("invalid self IP address: %s", config.SelfIp)
		}
		_, subnet, err := net.ParseCIDR(config.Subnet)
		if err != nil {
			return nil, fmt.Errorf("failed to parse subnet: %v", err)
		}
		if !subnet.Contains(selfIP) {
			return nil, fmt.Errorf("self IP %s is not within subnet %s", config.SelfIp, config.Subnet)
		}

		// If SelfIP is specified, configure network interface
		log.Logger.Debugf("Configuring interface %s with IP %s", config.DhcpServerInterface, config.SelfIp)
		if err := configureInterface(config.DhcpServerInterface, config.SelfIp); err != nil {
			log.Logger.Errorf("failed to configure interface: %v", err)
			return nil, err
		}

	} else {
		// Get interface addresses
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, fmt.Errorf("failed to get addresses for interface %s: %v", config.DhcpServerInterface, err)
		}

		// Parse the configured subnet
		_, configuredSubnet, err := net.ParseCIDR(config.Subnet)
		if err != nil {
			return nil, fmt.Errorf("invalid subnet %s: %v", config.Subnet, err)
		}

		// Check if any interface IP matches the configured subnet
		var hasMatchingIP bool
		for _, addr := range addrs {
			// Convert addr to IPNet
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			// Check if the interface IP is in the same subnet
			if ipNet.Mask.String() == configuredSubnet.Mask.String() &&
				configuredSubnet.Contains(ipNet.IP) {
				hasMatchingIP = true
				break
			}
		}

		if !hasMatchingIP {
			return nil, fmt.Errorf("interface %s has no IP address in subnet %s", config.DhcpServerInterface, config.Subnet)
		}
	}

	// Initialize dhcp server
	server := &dhcpServer{
		config:           config,
		clusterAgentName: clusterAgentName,
		stopChan:         make(chan struct{}),
		leaseFilePath:    fmt.Sprintf(DhcpLeaseFileFormat, clusterAgentName),
		addChan:          addChan,
		deleteChan:       deleteChan,
	}

	return server, nil
}

// Start initializes and starts the DHCP server.
// It performs the following steps:
// 1. Configures the network interface
// 2. Calculates total available IPs
// 3. Generates DHCP configuration
// 4. Starts the DHCP daemon
// 5. Begins monitoring routine
//
// Returns an error if any step fails.
func (s *dhcpServer) Start() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	log.Logger.Debugf("Starting DHCP server with config: interface=%s, subnet=%s, range=%s",
		s.config.DhcpServerInterface, s.config.Subnet, s.config.IpRange)

	// Calculate total IPs at startup
	if err := s.calculateTotalIPs(); err != nil {
		s.ipRangeErr = err
		return fmt.Errorf("failed to calculate total IPs: %v", err)
	}

	// Ensure DHCP configuration directory exists
	dhcpConfigDir := filepath.Dir(DhcpConfigPath)
	if err := os.MkdirAll(dhcpConfigDir, 0755); err != nil {
		return fmt.Errorf("failed to create DHCP configuration directory: %v", err)
	}

	// Generate dhcpd configuration
	if err := s.generateConfig(); err != nil {
		log.Logger.Errorf("failed to generate DHCP configuration: %v", err)
		return err
	}

	// Check if lease file exists and create if it doesn't
	_, err := os.Stat(s.leaseFilePath)
	if err == nil {
		log.Logger.Infof("Found existing DHCP lease file, will use it: %s", s.leaseFilePath)
	} else if os.IsNotExist(err) {
		// Create lease file if it doesn't exist
		leaseDir := filepath.Dir(s.leaseFilePath)
		if err := os.MkdirAll(leaseDir, 0755); err != nil {
			return fmt.Errorf("failed to create lease directory: %v", err)
		}

		if err := os.WriteFile(s.leaseFilePath, []byte(""), 0644); err != nil {
			return fmt.Errorf("failed to create lease file: %v", err)
		}
	} else {
		return fmt.Errorf("failed to check lease file: %v", err)
	}

	// Ensure log directory exists
	logDir := filepath.Dir(DhcpLogFile)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %v", err)
	}

	log.Logger.Debugf("starting DHCP server with command: %s", DhcpBinary)
	cmdArgs := []string{
		"-f",                  // Run in foreground
		"-cf", DhcpConfigPath, // Config file
		"-lf", s.leaseFilePath, // Lease file
		"-pf", "/var/run/dhcpd.pid", // PID file
		"-tf", DhcpLogFile, // Log file
	}

	// Add debug flags if debug logging is enabled
	//if os.Getenv(EnvDhcpDebugLog) == "true" {
	cmdArgs = append(cmdArgs, "-d") // Debug mode
	log.Logger.Info("DHCP debug logging is enabled")
	//}

	cmdArgs = append(cmdArgs, s.config.DhcpServerInterface)
	log.Logger.Infof("starting DHCP server with command: %s %s", DhcpBinary, strings.Join(cmdArgs, " "))

	s.cmd = exec.Command(DhcpBinary, cmdArgs...)

	// Set up logging to both file and our logger
	logFile, err := os.OpenFile(DhcpLogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}

	s.cmd.Stdout = logFile
	s.cmd.Stderr = logFile

	// Start the DHCP server
	if err := s.cmd.Start(); err != nil {
		logFile.Close()
		return fmt.Errorf("failed to start DHCP server: %v", err)
	}

	// Start a goroutine to close the log file when the process exits
	go func() {
		if err := s.cmd.Wait(); err != nil {
			log.Logger.Errorf("DHCP server process exited with error: %v", err)
			// Attempt to restart the server after a brief delay
			time.Sleep(5 * time.Second)
			if err := s.Start(); err != nil {
				log.Logger.Errorf("Failed to restart DHCP server: %v", err)
			} else {
				log.Logger.Info("DHCP server restarted successfully")
			}
		}
		logFile.Close()
	}()

	log.Logger.Info("DHCP server started successfully")

	// Start monitoring routine
	go s.monitor()

	return nil
}

// Stop gracefully stops the DHCP server.
// It performs the following steps:
// 1. Signals the monitoring routine to stop
// 2. Terminates the DHCP daemon process
// 3. Cleans up resources
//
// Returns an error if the server cannot be stopped.
func (s *dhcpServer) Stop() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	log.Logger.Debugf("Stopping DHCP server...")

	if s.cmd == nil || s.cmd.Process == nil {
		log.Logger.Debug("DHCP server is not running")
		return nil
	}

	// Signal monitor to stop
	close(s.stopChan)

	log.Logger.Debugf("Terminating DHCP server process (PID: %d)", s.cmd.Process.Pid)
	if err := s.cmd.Process.Kill(); err != nil {
		log.Logger.Errorf("failed to stop DHCP server: %v", err)
		return err
	}
	if err := s.cmd.Wait(); err != nil {
		log.Logger.Errorf("failed to wait for DHCP server: %v", err)
	}
	s.cmd = nil
	log.Logger.Info("DHCP server stopped successfully")

	return nil
}

// GetClientInfo returns information about DHCP clients
func (s *dhcpServer) GetClientInfo() ([]types.ClientInfo, error) {
	log.Logger.Debugf("Retrieving DHCP client information from lease file: %s", s.leaseFilePath)
	clients, err := GetDhcpClients(s.leaseFilePath)
	if err != nil {
		log.Logger.Debugf("Failed to get DHCP clients: %v", err)
		return nil, err
	}
	log.Logger.Debugf("Found %d DHCP clients", len(clients))
	return clients, nil
}
