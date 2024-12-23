package hoststatus

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/spidernet-io/bmc/pkg/agent/hoststatus/data"
	"github.com/spidernet-io/bmc/pkg/log"
	"github.com/spidernet-io/bmc/pkg/redfish"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var hostStatusLock sync.RWMutex

func (c *hostStatusController) UpdateHostStatusCr(d *data.HostConnectCon) error {

	// 创建 redfish 客户端
	client := redfish.NewClient(*d)

	protocol := "http"
	if d.Info.Https {
		protocol = "https"
	}
	auth := "without username and password"
	if len(d.Username) != 0 && len(d.Password) != 0 {
		auth = "with username and password"
	}
	log.Logger.Debugf("try to check redfish with url: %s://%s:%d , %s", protocol, d.Info.IpAddr, d.Info.Port, auth)

	// 获取 HostStatus 名称
	name := formatHostStatusName(c.config.ClusterAgentName, d.Info.IpAddr)
	// 获取现有的 HostStatus
	existing, err := c.client.HostStatuses().Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		log.Logger.Errorf("Failed to get HostStatus %s: %v", name, err)
		return err
	}
	updated := existing.DeepCopy()

	// 检查健康状态
	healthy := client.Health()
	updated.Status.HealthReady = healthy
	client.GetInfo()

	// 更新 HostStatus
	if !reflect.DeepEqual(updated.Status, existing.Status) {
		updated.Status.LastUpdateTime = time.Now().UTC().Format(time.RFC3339)
		_, err = c.client.HostStatuses().UpdateStatus(context.Background(), updated, metav1.UpdateOptions{})
		if err != nil {
			log.Logger.Errorf("Failed to update status of HostStatus %s: %v", name, err)
			return err
		}
		log.Logger.Infof("Successfully updated HostStatus %s status", name)
	} else {
		log.Logger.Debugf("no need to updated HostStatus %s status")
	}

	return nil
}

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

	for name, t := range syncData {
		log.Logger.Debugf("update status of the hostStatus %s ", name)
		c.UpdateHostStatusCr(&t)
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
