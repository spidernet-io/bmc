package redfish

import (
	"fmt"

	"github.com/stmcginnis/gofish"
)

// Health 实现健康检查方法
func (c *redfishClient) Health() bool {
	// 创建 gofish 客户端
	client, err := gofish.Connect(c.config)
	if err != nil {
		return false
	}
	defer client.Logout()
	return true
}

// Health 实现健康检查方法
func (c *redfishClient) GetInfo() error {

	// 创建 gofish 客户端
	client, err := gofish.Connect(c.config)
	if err != nil {
		c.logger.Errorf("failed to connect: %+v", err)
		return err
	}
	defer client.Logout()

	// Attached the client to service root
	service := client.Service

	c.logger.Debugf("RedfishVersion: %v", service.RedfishVersion)
	c.logger.Debugf("Vendor: %v", service.Vendor)

	// Query the computer systems
	ss, err := service.Systems()
	if err != nil {
		c.logger.Errorf("failed to Query the computer systems: %+v", err)
		return err
	} else if len(ss) == 0 {
		c.logger.Errorf("failed to get system")
		return fmt.Errorf("failed to get system")
	}
	c.logger.Debugf("system amount: %d", len(ss))
	for n, t := range ss {
		c.logger.Debugf("systems[%d]: %+v", n, *t)
	}
	// for barel metal case,
	system := ss[0]
	c.logger.Debugf("BiosVerison: %v", system.BIOSVersion)
	c.logger.Debugf("HostName: %v", system.HostName)
	c.logger.Debugf("Manufacturer: %v", system.Manufacturer)
	c.logger.Debugf("MemoryGiB: %v", system.MemorySummary.TotalSystemMemoryGiB)
	c.logger.Debugf("CpuPhysicalCore: %v", system.ProcessorSummary.Count)
	c.logger.Debugf("PowerState: %v", system.PowerState)
	c.logger.Debugf("Status: %v", system.Status.Health)

	// Query the managers for bmc
	managers, err := service.Managers()
	if err != nil {
		c.logger.Errorf("failed to Query the bmc : %+v", err)
		return err
	} else if len(managers) == 0 {
		c.logger.Errorf("failed to get bmc")
		return fmt.Errorf("failed to get bmc")
	}
	c.logger.Debugf("bmc amount: %d", len(managers))
	for n, t := range managers {
		c.logger.Debugf("bmc[%d]: %+v", n, *t)
	}
	bmc := managers[0]
	c.logger.Debugf("BmcFirmwareVersion: %v", bmc.FirmwareVersion)

	return nil
}
