package hoststatus

import (
	"sync"

	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/spidernet-io/bmc/pkg/agent/config"
	"github.com/spidernet-io/bmc/pkg/dhcpserver/types"
	bmcv1beta1 "github.com/spidernet-io/bmc/pkg/k8s/apis/bmc.spidernet.io/v1beta1"
	"github.com/spidernet-io/bmc/pkg/log"
)

type HostStatusController interface {
	GetDHCPEventChan() (chan<- types.ClientInfo, chan<- types.ClientInfo)
	Stop()
	SetupWithManager(mgr ctrl.Manager) error
}

type hostStatusController struct {
	client     client.Client
	kubeClient kubernetes.Interface
	config     *config.AgentConfig
	addChan    chan types.ClientInfo
	deleteChan chan types.ClientInfo
	stopCh     chan struct{}
	wg         sync.WaitGroup
}

func NewHostStatusController(kubeClient kubernetes.Interface, config *config.AgentConfig, mgr ctrl.Manager) HostStatusController {
	log.Logger.Debugf("Creating new HostStatus controller for cluster agent: %s", config.ClusterAgentName)
	controller := &hostStatusController{
		client:     mgr.GetClient(),
		kubeClient: kubeClient,
		config:     config,
		addChan:    make(chan types.ClientInfo),
		deleteChan: make(chan types.ClientInfo),
		stopCh:     make(chan struct{}),
	}

	log.Logger.Debugf("HostStatus controller created successfully")
	return controller
}

func (c *hostStatusController) Stop() {
	log.Logger.Info("Stopping HostStatus controller")
	close(c.stopCh)
	c.wg.Wait()
	log.Logger.Info("HostStatus controller stopped successfully")
}

func (c *hostStatusController) GetDHCPEventChan() (chan<- types.ClientInfo, chan<- types.ClientInfo) {
	return c.addChan, c.deleteChan
}

// SetupWithManager 设置 controller-runtime manager
func (c *hostStatusController) SetupWithManager(mgr ctrl.Manager) error {

	// 启动 DHCP 事件处理
	go c.processDHCPEvents()

	// 启动 hoststatus spec.info 的	周期更新
	go c.UpdateHostStatusAtInterval()

	return ctrl.NewControllerManagedBy(mgr).
		For(&bmcv1beta1.HostStatus{}).
		Complete(c)
}
