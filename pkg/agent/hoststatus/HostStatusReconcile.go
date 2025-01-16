package hoststatus

import (
	"context"
	"fmt"
	"sync"
	"time"

	hoststatusdata "github.com/spidernet-io/bmc/pkg/agent/hoststatus/data"
	bmcv1beta1 "github.com/spidernet-io/bmc/pkg/k8s/apis/bmc.spidernet.io/v1beta1"

	//"github.com/spidernet-io/bmc/pkg/lock"
	"github.com/spidernet-io/bmc/pkg/log"
	"github.com/spidernet-io/bmc/pkg/redfish"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

// the lock-holding timeout is long because it needs to send http request to redfish for each host
// so it uses sync.Mutex instead of lock.Mutex
var hostStatusLock = &sync.Mutex{}

// ------------------------------  update the spec.info of the hoststatus

// this is called by UpdateHostStatusAtInterval and UpdateHostStatusWrapper
func (c *hostStatusController) UpdateHostStatusInfo(name string, d *hoststatusdata.HostConnectCon) (bool, error) {

	// local lock for updateing each hostStatus
	hostStatusLock.Lock()
	defer hostStatusLock.Unlock()

	// 创建 redfish 客户端
	var healthy bool
	client, err1 := redfish.NewClient(*d, log.Logger)
	if err1 != nil {
		log.Logger.Errorf("Failed to create redfish client for HostStatus %s: %v", name, err1)
		healthy = false
	} else {
		healthy = true
	}

	protocol := "http"
	if d.Info.Https {
		protocol = "https"
	}
	auth := "without username and password"
	if len(d.Username) != 0 && len(d.Password) != 0 {
		auth = "with username and password"
	}
	log.Logger.Debugf("try to check redfish with url: %s://%s:%d , %s", protocol, d.Info.IpAddr, d.Info.Port, auth)

	// 获取现有的 HostStatus
	existing := &bmcv1beta1.HostStatus{}
	err := c.client.Get(context.Background(), types.NamespacedName{Name: name}, existing)
	if err != nil {
		log.Logger.Errorf("Failed to get HostStatus %s: %v", name, err)
		return false, err
	}
	updated := existing.DeepCopy()

	// 检查健康状态
	updated.Status.Healthy = healthy
	if healthy {
		infoData, err := client.GetInfo()
		if err != nil {
			log.Logger.Errorf("Failed to get info of HostStatus %s: %v", name, err)
			healthy = false
		} else {
			updated.Status.Info = infoData
		}
	}
	if !healthy {
		log.Logger.Debugf("HostStatus %s is not healthy, set info to empty", name)
		updated.Status.Info = map[string]string{}
	}
	if updated.Status.Healthy != existing.Status.Healthy {
		log.Logger.Infof("HostStatus %s change from %v to %v , update status", name, existing.Status.Healthy, healthy)
	}

	// 更新 HostStatus
	if !compareHostStatus(updated.Status, existing.Status, log.Logger) {
		log.Logger.Debugf("status changed, existing: %v, updated: %v", existing.Status, updated.Status)
		updated.Status.LastUpdateTime = time.Now().UTC().Format(time.RFC3339)
		if err := c.client.Status().Update(context.Background(), updated); err != nil {
			log.Logger.Errorf("Failed to update status of HostStatus %s: %v", name, err)
			return true, err
		}
		log.Logger.Infof("Successfully updated HostStatus %s status", name)
		return true, nil
	}
	return false, nil
}

// this is called by UpdateHostStatusAtInterval and
func (c *hostStatusController) UpdateHostStatusInfoWrapper(name string) error {
	syncData := make(map[string]hoststatusdata.HostConnectCon)
	modeinfo := ""
	if len(name) == 0 {
		syncData = hoststatusdata.HostCacheDatabase.GetAll()
		if len(syncData) == 0 {
			return nil
		}
		modeinfo = " during periodic update"
	} else {
		d := hoststatusdata.HostCacheDatabase.Get(name)
		if d != nil {
			syncData[name] = *d
		}
		if len(syncData) == 0 {
			log.Logger.Errorf("no cache data found for hostStatus %s ", name)
			return fmt.Errorf("no cache data found for hostStatus %s ", name)
		}
		modeinfo = " during hoststatus reconcile"
	}

	for item, t := range syncData {
		if updated, err := c.UpdateHostStatusInfo(item, &t); err != nil {
			log.Logger.Errorf("failed to update HostStatus %s %s: %v", item, modeinfo, err)
		} else {
			if updated {
				log.Logger.Debugf("update status of the hostStatus %s %s", item, modeinfo)
			} else {
				log.Logger.Debugf("no need to update status of the hostStatus %s %s", item, modeinfo)
			}
		}
	}

	return nil
}

// ------------------------------  hoststatus spec.info 的	周期更新
func (c *hostStatusController) UpdateHostStatusAtInterval() {
	interval := time.Duration(c.config.HostStatusUpdateInterval) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	log.Logger.Infof("begin to update all hostStatus at interval of %v seconds", c.config.HostStatusUpdateInterval)

	for {
		select {
		case <-c.stopCh:
			log.Logger.Info("Stopping UpdateHostStatusAtInterval")
			return
		case <-ticker.C:
			log.Logger.Debugf("update all hostStatus at interval ")
			if err := c.UpdateHostStatusInfoWrapper(""); err != nil {
				log.Logger.Errorf("Failed to update host status: %v", err)
			}
		}
	}
}

// ------------------------------  hoststatus 的 reconcile , 触发更新

// 缓存 hostStatus 数据本地，并行更新 status.info 信息
func (c *hostStatusController) processHostStatus(hostStatus *bmcv1beta1.HostStatus, logger *zap.SugaredLogger) error {

	logger.Debugf("Processing Existed HostStatus: %s (Type: %s, IP: %s, Health: %v)",
		hostStatus.Name,
		hostStatus.Status.Basic.Type,
		hostStatus.Status.Basic.IpAddr,
		hostStatus.Status.Healthy)

	// cache the hostStatus data to local
	username, password, err := c.getSecretData(
		hostStatus.Status.Basic.SecretName,
		hostStatus.Status.Basic.SecretNamespace,
	)
	if err != nil {
		logger.Errorf("Failed to get secret data for HostStatus %s: %v", hostStatus.Name, err)
		return err
	}

	logger.Debugf("Adding/Updating HostStatus %s in cache with username: %s",
		hostStatus.Name, username)

	hoststatusdata.HostCacheDatabase.Add(hostStatus.Name, hoststatusdata.HostConnectCon{
		Info:     &hostStatus.Status.Basic,
		Username: username,
		Password: password,
		DhcpHost: hostStatus.Status.Basic.Type == bmcv1beta1.HostTypeDHCP,
	})

	if len(hostStatus.Status.Info) == 0 {
		if err := c.UpdateHostStatusInfoWrapper(hostStatus.Name); err != nil {
			logger.Errorf("failed to update HostStatus %s: %v", hostStatus.Name, err)
			return err
		}
	} else {
		logger.Debugf("HostStatus %s has already been processed, skipping the first time update", hostStatus.Name)
	}

	logger.Debugf("Successfully processed HostStatus %s", hostStatus.Name)
	return nil
}

// Reconcile 实现 reconcile.Reconciler 接口
// 负责在 hoststatus 创建后 Info 信息的第一次更新（后续的更新由 UpdateHostStatusAtInterval 完成）
func (c *hostStatusController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.Logger.With(
		zap.String("hoststatus", req.Name),
	)

	logger.Debugf("Reconciling HostStatus %s", req.Name)

	// 获取 HostStatus
	hostStatus := &bmcv1beta1.HostStatus{}
	if err := c.client.Get(ctx, req.NamespacedName, hostStatus); err != nil {
		if errors.IsNotFound(err) {
			logger.Debugf("HostStatus not found, delete from cache")
			hoststatusdata.HostCacheDatabase.Delete(req.Name)
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get HostStatus")
		return ctrl.Result{}, err
	}

	if hostStatus.Status.ClusterAgent == "" {
		logger.Debugf("HostStatus %s has no clusterAgent, skipping", hostStatus.Name)
		return ctrl.Result{}, nil
	}

	if hostStatus.Status.ClusterAgent != c.config.ClusterAgentName {
		logger.Debugf("HostStatus %s belongs to another agent %s, skipping", hostStatus.Name, hostStatus.Status.ClusterAgent)
		return ctrl.Result{}, nil
	}

	if len(hostStatus.Status.Basic.IpAddr) == 0 {
		// the hostStatus is created firstly and then be updated with its status
		log.Logger.Debugf("ingore hostStatus %s just created", hostStatus.Name)
		return ctrl.Result{}, nil
	}

	// 处理 HostStatus
	if err := c.processHostStatus(hostStatus, logger); err != nil {
		logger.Error(err, "Failed to process HostStatus, will retry")
		return ctrl.Result{
			RequeueAfter: time.Second * 2,
		}, err
	}

	logger.Debugf("Successfully processed HostStatus %s", hostStatus.Name)
	return ctrl.Result{}, nil
}
