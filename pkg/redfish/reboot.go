package redfish

import (
	"fmt"

	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
)

type BootCmd string

const (
	BootCmdOn           BootCmd = "on"
	BootCmdOff          BootCmd = "off"
	BootCmdReset        BootCmd = "reset"
	BootCmdResetPxeOnce BootCmd = "resetPxeOnce"
)

// https://github.com/stmcginnis/gofish/blob/main/examples/reboot.md
// https://github.com/DMTF/Redfish-Tacklebox/blob/main/scripts/rf_power_reset.py
// post request to systems
/*
# curl -u "" -k https://10.64.64.94/redfish/v1/Systems/1/ResetActionInfo
{
  "@odata.type": "#ActionInfo.v1_1_2.ActionInfo",
  "@odata.id": "/redfish/v1/Systems/1/ResetActionInfo",
  "Id": "ResetActionInfo",
  "Name": "Reset Action Info",
  "Parameters": [
    {
      "Name": "ResetType",
      "Required": true,
      "DataType": "String",
      "AllowableValues": [
        "On",
        "ForceOff",
        "GracefulShutdown",
        "GracefulRestart",
        "ForceRestart",
        "Nmi",
        "ForceOn",
        "PowerCycle"
      ]
    }
  ],
  "Oem": {}
}
*/
func (c *redfishClient) Reboot(bootCmd BootCmd) error {

	// 创建 gofish 客户端
	client, err := gofish.Connect(c.config)
	if err != nil {
		c.logger.Errorf("failed to connect: %+v", err)
		return err
	}
	defer client.Logout()

	// Attached the client to service root
	service := client.Service
	// Query the computer systems
	ss, err := service.Systems()
	if err != nil {
		c.logger.Errorf("failed to Query the computer systems: %+v", err)
		return err
	}

	// Creates a boot override to pxe once
	bootOverride := redfish.Boot{
		// boot from the Pre-Boot EXecution (PXE) environment
		BootSourceOverrideTarget: redfish.PxeBootSourceOverrideTarget,
		// boot (one time) to the Boot Source Override Target
		BootSourceOverrideEnabled: redfish.OnceBootSourceOverrideEnabled,
	}

	for _, system := range ss {

		switch bootCmd {
		case BootCmdOn:
			err = system.Reset(redfish.ForceOnResetType)
		case BootCmdOff:
			err = system.Reset(redfish.ForceOffResetType)
		case BootCmdReset:
			if bootCmd == BootCmdResetPxeOnce {
				c.logger.Debugf("pxe reboot for System: %+v \n", system)
				err := system.SetBoot(bootOverride)
				if err != nil {
					return fmt.Errorf("failed to set boot option")
				}
			} else {
				c.logger.Debugf("normal reboot for System: %+v \n", system)
			}
			err = system.Reset(redfish.ForceRestartResetType)
		}
		if err != nil {
			c.logger.Errorf("failed to operate system %+v: %+v \n", system, err)
			return fmt.Errorf("failed to operate ")
		}
	}

	return nil
}
