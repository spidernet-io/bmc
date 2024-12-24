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

func setData(result map[string]string, key, value string) {
	if len(value) == 0 {
		result[key] = ""
	} else {
		result[key] = value
	}
	return
}

// Health 实现健康检查方法
func (c *redfishClient) GetInfo() (map[string]string, error) {

	result := map[string]string{}

	// 创建 gofish 客户端
	client, err := gofish.Connect(c.config)
	if err != nil {
		c.logger.Errorf("failed to connect: %+v", err)
		return nil, err
	}
	defer client.Logout()

	// Attached the client to service root
	service := client.Service

	setData(result, "RedfishVersion", service.RedfishVersion)
	setData(result, "Vendor", service.Vendor)

	// Query the computer systems
	ss, err := service.Systems()
	if err != nil {
		c.logger.Errorf("failed to Query the computer systems: %+v", err)
		return nil, err
	} else if len(ss) == 0 {
		c.logger.Errorf("failed to get system")
		return nil, fmt.Errorf("failed to get system")
	}
	c.logger.Debugf("system amount: %d", len(ss))
	for n, t := range ss {
		c.logger.Debugf("systems[%d]: %+v", n, *t)
	}
	// for barel metal case,
	system := ss[0]
	setData(result, "BiosVerison", system.BIOSVersion)
	setData(result, "HostName", system.HostName)
	setData(result, "Manufacturer", system.Manufacturer)
	setData(result, "MemoryGiB", fmt.Sprintf("%f", system.MemorySummary.TotalSystemMemoryGiB))
	setData(result, "CpuPhysicalCore", fmt.Sprintf("%d", system.ProcessorSummary.Count))
	setData(result, "CpuLogicalCore", fmt.Sprintf("%d", system.ProcessorSummary.LogicalProcessorCount))
	setData(result, "PowerState", string(system.PowerState))
	setData(result, "SyatemStatus", string(system.Status.Health))

	// optional: in old redfish version, the following fields are missing
	pcieList, err := system.PCIeDevices()
	if err == nil && len(pcieList) > 0 {
		c.logger.Debugf("PCIeDevices: %v", pcieList)
	}

	// Query the managers for bmc
	managers, err := service.Managers()
	if err != nil {
		c.logger.Errorf("failed to Query the bmc : %+v", err)
		return nil, err
	} else if len(managers) == 0 {
		c.logger.Errorf("failed to get bmc")
		return nil, fmt.Errorf("failed to get bmc")
	}
	c.logger.Debugf("bmc amount: %d", len(managers))
	for n, t := range managers {
		c.logger.Debugf("bmc[%d]: %+v", n, *t)
	}
	bmc := managers[0]
	setData(result, "BmcFirmwareVersion", bmc.FirmwareVersion)
	setData(result, "BmcStatus", string(bmc.Status.Health))

	return result, nil
}
