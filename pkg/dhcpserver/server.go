// Package dhcpserver provides DHCP server management functionality.
package dhcpserver

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"path/filepath"

	bmcv1beta1 "github.com/spidernet-io/bmc/pkg/apis/bmc/v1beta1"
	"github.com/spidernet-io/bmc/pkg/log"
)

// dhcpServer implements the DhcpServer interface.
// It manages the lifecycle of an ISC DHCP server instance and provides
// methods to monitor its operation and retrieve statistics.
type dhcpServer struct {
	// config holds the DHCP server configuration
	config *bmcv1beta1.DhcpServerConfig
	// cmd represents the running DHCP server process
	cmd *exec.Cmd
	// mutex protects access to shared resources
	mutex sync.Mutex
	// stopChan signals the monitoring routine to stop
	stopChan chan struct{}
	// stats holds current IP usage statistics
	stats IPUsageStats
	// totalIPs is the total number of IP addresses available for allocation
	totalIPs int
	// ipRangeErr stores any error encountered while parsing IP range
	ipRangeErr error
}

// NewDhcpServer creates a new DHCP server instance.
// Parameters:
//   - config: DHCP server configuration including interface, subnet, and IP range
//
// Returns:
//   - DhcpServer interface implementation
func NewDhcpServer(config *bmcv1beta1.DhcpServerConfig) DhcpServer {
	return &dhcpServer{
		config:   config,
		stopChan: make(chan struct{}),
	}
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

	// If SelfIP is specified, configure network interface
	if s.config.SelfIp != "" {
		log.Logger.Debugf("Configuring interface %s with IP %s", s.config.DhcpServerInterface, s.config.SelfIp)
		if err := s.configureInterface(); err != nil {
			log.Logger.Errorf("failed to configure interface: %v", err)
			return err
		}
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

	// Check if DHCP lease file exists, if so, use it
	leaseFileStat, err := os.Stat(DhcpLeaseFile)
	if err == nil && leaseFileStat.Mode().IsRegular() {
		log.Logger.Infof("Found existing DHCP lease file, will use it: %s", DhcpLeaseFile)
	} else {
		log.Logger.Infof("No existing DHCP lease file found, will start from scratch")
		// Ensure lease directory exists
		leaseDir := filepath.Dir(DhcpLeaseFile)
		if _, err := os.Stat(leaseDir); os.IsNotExist(err) {
			if err := os.MkdirAll(leaseDir, 0755); err != nil {
				return fmt.Errorf("failed to create lease directory: %v", err)
			}
		}
		// Create empty lease file if it doesn't exist
		log.Logger.Info("Creating empty DHCP lease file")
		if err := os.WriteFile(DhcpLeaseFile, []byte(""), 0644); err != nil {
			return fmt.Errorf("failed to create lease file: %v", err)
		}
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
		"-lf", DhcpLeaseFile, // Lease file
		"-pf", "/var/run/dhcpd.pid", // PID file
		"-tf", DhcpLogFile, // Log file
	}

	// Add debug flags if debug logging is enabled
	//if os.Getenv(EnvDhcpDebugLog) == "true" {
	cmdArgs = append(cmdArgs, "-d") // Debug mode
	log.Logger.Info("DHCP debug logging is enabled")
	//}

	cmdArgs = append(cmdArgs, s.config.DhcpServerInterface)
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

	// Initialize IP usage stats
	if err := s.updateStats(); err != nil {
		log.Logger.Warnf("Failed to initialize IP usage stats: %v", err)
	}

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

	log.Logger.Debugf("Stopping DHCP server on interface %s", s.config.DhcpServerInterface)

	if s.cmd != nil && s.cmd.Process != nil {
		log.Logger.Debug("stopping DHCP server process")
		close(s.stopChan)
		if err := s.cmd.Process.Kill(); err != nil {
			log.Logger.Errorf("failed to stop DHCP server: %v", err)
			return err
		}
		s.cmd.Wait()
		s.cmd = nil
		log.Logger.Info("DHCP server stopped successfully")
	}

	return nil
}

// GetClientInfo returns information about DHCP clients
func (s *dhcpServer) GetClientInfo() ([]ClientInfo, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	log.Logger.Debug("Retrieving DHCP client information")
	return parseLeasesFile()
}

// GetIPUsageStats calculates current IP allocation statistics.
// It counts the number of active leases and compares with total
// available IPs to determine usage statistics.
//
// Returns:
//   - *IPUsageStats: current IP usage statistics
//   - error: if statistics cannot be calculated
func (s *dhcpServer) GetIPUsageStats() (*IPUsageStats, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	log.Logger.Debugf("Retrieving IP usage statistics")
	if err := s.updateStats(); err != nil {
		return nil, err
	}
	return &s.stats, nil
}

// monitor periodically checks server health and updates statistics.
// This routine runs in a separate goroutine and continues until
// signaled to stop via stopChan.
func (s *dhcpServer) monitor() {
	ticker := time.NewTicker(time.Duration(MonitorInterval) * time.Second)
	defer ticker.Stop()

	log.Logger.Debugf("Starting DHCP server monitor with interval %d seconds", MonitorInterval)

	for {
		select {
		case <-s.stopChan:
			log.Logger.Debugf("DHCP server monitor stopped")
			return
		case <-ticker.C:
			if s.cmd == nil || s.cmd.Process == nil {
				log.Logger.Infof("DHCP server not running, attempting restart...")
				s.Start()
			}
			// Update IP usage stats
			if err := s.updateStats(); err != nil {
				log.Logger.Warnf("Failed to update IP usage stats: %v", err)
			}
		}
	}
}

// updateStats calculates current IP allocation statistics.
// It counts the number of active leases and compares with total
// available IPs to determine usage statistics.
//
// Returns an error if statistics cannot be calculated.
func (s *dhcpServer) updateStats() error {
	// If there was an error parsing IP range, return it
	if s.ipRangeErr != nil {
		return s.ipRangeErr
	}

	clients, err := parseLeasesFile()
	if err != nil {
		log.Logger.Errorf("failed to parse lease file: %v", err)
		return err
	}

	// Calculate new stats
	newStats := IPUsageStats{
		TotalIPs:     s.totalIPs,
		AvailableIPs: s.totalIPs - len(clients),
	}

	// Only log if stats have changed
	if newStats.AvailableIPs != s.stats.AvailableIPs {
		log.Logger.Debugf("IP usage stats updated - Total: %d, Available: %d", newStats.TotalIPs, newStats.AvailableIPs)
	}

	// Update stats
	s.stats = newStats
	return nil
}
