package hoststatus

import (
	"fmt"
	"sync"
	"time"

	"github.com/spidernet-io/bmc/pkg/agent/hoststatus/data"
	"github.com/spidernet-io/bmc/pkg/log"
)

var hostStatusLock sync.RWMutex

func (c *hostStatusController) UpdateHostStatus(name string) error {
	syncData := make(map[string]data.HostConnectCon)

	if len(name) == 0 {
		syncData = data.HostCacheDatabase.GetAll()
		if len(syncData) == 0 {
			return nil
		}
	} else {
		d := data.HostCacheDatabase.Get(name)
		if d != nil {
			syncData[name] = *d
		}
		if len(syncData) == 0 {
			log.Logger.Errorf("no cache data found for hostStatus %s ", name)
			return fmt.Errorf("no cache data found for hostStatus %s ", name)
		}
	}

	hostStatusLock.Lock()
	defer hostStatusLock.Unlock()

	for name, _ := range syncData {
		log.Logger.Debugf("update status of the hostStatus %s ", name)
	}

	return nil
}

func (c *hostStatusController) UpdateHostStatusAtInterval() {

	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	log.Logger.Infof("begin to update all hostStatus at interval of %v seconds", 60)

	for {
		select {
		case <-c.stopCh: // 使用 controller 中的 stopCh 来控制退出
			log.Logger.Info("Stopping UpdateHostStatusAtInterval")
			return
		case <-ticker.C:
			log.Logger.Debugf("update all hostStatus at interval ")
			if err := c.UpdateHostStatus(""); err != nil {
				log.Logger.Errorf("Failed to update host status: %v", err)
			}
		}
	}
}
