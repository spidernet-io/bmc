package redfish

import (
	"fmt"

	bmcv1beta1 "github.com/spidernet-io/bmc/pkg/k8s/apis/bmc.spidernet.io/v1beta1"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
)

// https://github.com/DMTF/Redfish-Tacklebox/blob/main/scripts/rf_power_reset.py
// post request to systems

func (c *redfishClient) Power(bootCmd string) error {

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
	if len(ss) == 0 {
		c.logger.Errorf("no system found")
		return fmt.Errorf("no system found")
	}

	for _, system := range ss {
		bootOptions, err := system.BootOptions()
		if err != nil {
			c.logger.Errorf("failed to get boot options: %+v", err)
			return err
		}
		c.logger.Debugf("system %s, boot options: %+v", system.Name, bootOptions)
		c.logger.Debugf("system %s, boot : %+v", system.Name, system.Boot)
		c.logger.Debugf("system %s, supported reset types: %+v", system.Name, system.SupportedResetTypes)

		switch bootCmd {
		case bmcv1beta1.BootCmdOn:
			fallthrough
		case bmcv1beta1.BootCmdForceOn:
			fallthrough
		case bmcv1beta1.BootCmdForceOff:
			fallthrough
		case bmcv1beta1.BootCmdGracefulShutdown:
			fallthrough
		case bmcv1beta1.BootCmdForceRestart:
			fallthrough
		case bmcv1beta1.BootCmdGracefulRestart:
			c.logger.Infof("operation %s on %s for System: %+v \n", bootCmd, c.config.Endpoint, system.Name)
			err = system.Reset(redfish.ResetType(bootCmd))

		case bmcv1beta1.BootCmdResetPxeOnce:
			// https://github.com/stmcginnis/gofish/blob/main/examples/reboot.md
			// Creates a boot override to pxe once
			bootOverride := redfish.Boot{
				// boot from the Pre-Boot EXecution (PXE) environment
				BootSourceOverrideTarget: redfish.PxeBootSourceOverrideTarget,
				// boot (one time) to the Boot Source Override Target
				BootSourceOverrideEnabled: redfish.OnceBootSourceOverrideEnabled,
			}
			c.logger.Infof("pxe reboot %s for System: %+v \n", c.config.Endpoint, system.Name)
			err := system.SetBoot(bootOverride)
			if err != nil {
				return fmt.Errorf("failed to set boot option")
			}
			err = system.Reset(redfish.ForceRestartResetType)

		default:
			c.logger.Errorf("unknown boot cmd: %+v", bootCmd)
			return fmt.Errorf("unknown boot cmd: %+v", bootCmd)
		}
		if err != nil {
			c.logger.Errorf("failed to operate system %+v: %+v \n", system, err)
			return fmt.Errorf("failed to operate ")
		}
	}

	return nil
}
