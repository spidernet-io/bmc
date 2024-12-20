package dhcpserver

import (
	"fmt"

	"github.com/spidernet-io/bmc/pkg/log"

	"syscall"
	"time"

	"github.com/spidernet-io/bmc/pkg/dhcpserver/types"
)

// GetIPUsageStats calculates current IP allocation statistics.
// It counts the number of active leases and compares with total
// available IPs to determine usage statistics.
//
// Returns:
//   - *IPUsageStats: current IP usage statistics
//   - error: if statistics cannot be calculated
func (s *dhcpServer) GetIPUsageStats() (*types.IPUsageStats, error) {

	log.Logger.Debugf("Retrieving IP usage statistics")
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if err := s.updateStats(); err != nil {
		return nil, err
	}
	return &s.stats, nil
}

// monitor periodically checks server health and updates statistics.
// This routine runs in a separate goroutine and continues until
// signaled to stop via stopChan.
func (s *dhcpServer) monitor() {
	ticker := time.NewTicker(MonitorInterval * time.Second)
	defer ticker.Stop()

	log.Logger.Debug("Initialize IP usage stats")
	if err := s.updateStats(); err != nil {
		log.Logger.Warnf("Failed to initialize IP usage stats: %v", err)
	}
	log.Logger.Debug("Starting DHCP server monitor routine")

	for {
		select {
		case <-s.stopChan:
			log.Logger.Debug("DHCP server monitor stopped")
			return
		case <-ticker.C:
			// Check if process exists and is running
			needRestart := s.cmd == nil || s.cmd.Process == nil
			if !needRestart {
				// Process exists, check if it's still running
				if err := s.cmd.Process.Signal(syscall.Signal(0)); err != nil {
					log.Logger.Debugf("DHCP server process check failed: %v", err)
					needRestart = true
				} else {
					log.Logger.Debug("DHCP server process health check passed")
				}
			}

			if needRestart {
				log.Logger.Warnf("DHCP server process not running, restarting...")

				// Print last 50 lines of log file before restart
				if err := s.printDhcpLogTail(); err != nil {
					log.Logger.Errorf("Failed to print DHCP log tail: %v", err)
				}

				if err := s.Start(); err != nil {
					log.Logger.Errorf("Failed to restart DHCP server: %v", err)
				}
			} else {
				if err := s.updateStats(); err != nil {
					log.Logger.Debugf("Failed to update stats during monitoring: %v", err)
				} else {
					log.Logger.Debugf("Successfully updated DHCP server stats - Available IPs: %d, Used: %d, Usage: %.2f%%",
						s.stats.AvailableIPs, s.stats.UsedIPs, s.stats.UsagePercentage)
				}
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

	//log.Logger.Debug("Updating DHCP server statistics")
	s.mutex.Lock()
	defer s.mutex.Unlock()

	clients, err := s.GetClientInfo()
	if err != nil {
		log.Logger.Errorf("Failed to get client info during stats update: %v", err)
		return fmt.Errorf("failed to get client info: %v", err)
	}

	// Compare with previous clients to detect changes
	newClients := make(map[string]types.ClientInfo)
	oldClients := make(map[string]types.ClientInfo)

	for _, client := range clients {
		newClients[client.IP] = client
	}
	for _, client := range s.previousClients {
		oldClients[client.IP] = client
	}

	if len(newClients) != len(oldClients) {
		log.Logger.Debugf("get GetClientInfo: %+v", clients)
		log.Logger.Infof("Checking for client changes - Current clients: %d, Previous clients: %d",
			len(newClients), len(oldClients))
	}

	// Check for new clients and MAC changes
	for ip, newClient := range newClients {
		oldClient, exists := oldClients[ip]
		if !exists {
			log.Logger.Infof("New DHCP client detected - IP: %s, MAC: %s", ip, newClient.MAC)
			// New IP allocation
			s.addChan <- newClient
		} else if oldClient.MAC != newClient.MAC {
			// Same IP but MAC changed, send delete followed by add
			log.Logger.Infof("DHCP client changed the mac - IP: %s, MAC: %s", ip, newClient.MAC)
			//if s.deleteChan != nil {
			//	s.deleteChan <- oldClient
			//}
			s.addChan <- newClient
		}
	}

	// Check for removed clients
	for ip, oldClient := range oldClients {
		if _, exists := newClients[ip]; !exists {
			log.Logger.Infof("Deleted DHCP client detected - IP: %s, MAC: %s", ip, oldClient.MAC)
			s.deleteChan <- oldClient
		}
	}

	s.previousClients = clients
	s.stats.UsedIPs = len(clients)
	s.stats.AvailableIPs = s.totalIPs - s.stats.UsedIPs
	if s.totalIPs > 0 {
		s.stats.UsagePercentage = float64(s.stats.UsedIPs) / float64(s.totalIPs) * 100
	}

	log.Logger.Debugf("DHCP server statistics updated - Available IPs: %d, Used: %d, Usage: %.2f%%",
		s.stats.AvailableIPs, s.stats.UsedIPs, s.stats.UsagePercentage)

	return nil
}
