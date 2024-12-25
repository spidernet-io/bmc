package hoststatus

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"k8s.io/client-go/kubernetes"

	"github.com/spidernet-io/bmc/pkg/agent/config"
	"github.com/spidernet-io/bmc/pkg/dhcpserver/types"
	bmcv1beta1 "github.com/spidernet-io/bmc/pkg/k8s/apis/bmc.spidernet.io/v1beta1"
	crdclientset "github.com/spidernet-io/bmc/pkg/k8s/client/clientset/versioned/typed/bmc.spidernet.io/v1beta1"
	"github.com/spidernet-io/bmc/pkg/log"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const (
	maxRetries = 10 // 最大重试次数
)

type HostStatusController interface {
	Run(ctx context.Context) error
	Stop()
	GetDHCPEventChan() (chan<- types.ClientInfo, chan<- types.ClientInfo)
}

type hostStatusController struct {
	client         *crdclientset.BmcV1beta1Client
	kubeClient     kubernetes.Interface
	config         *config.AgentConfig
	informer       cache.SharedIndexInformer
	statusInformer cache.SharedIndexInformer
	workqueue      workqueue.RateLimitingInterface
	addChan        chan types.ClientInfo
	deleteChan     chan types.ClientInfo
	stopCh         chan struct{}
	wg             sync.WaitGroup
}

func NewHostStatusController(client *crdclientset.BmcV1beta1Client, kubeClient kubernetes.Interface, config *config.AgentConfig) HostStatusController {
	log.Logger.Debugf("Creating new HostStatus controller for cluster agent: %s", config.ClusterAgentName)
	controller := &hostStatusController{
		client:     client,
		kubeClient: kubeClient,
		config:     config,
		workqueue:  workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		addChan:    make(chan types.ClientInfo),
		deleteChan: make(chan types.ClientInfo),
		stopCh:     make(chan struct{}),
	}

	log.Logger.Debugf("HostStatus controller created successfully")
	return controller
}

func (c *hostStatusController) Run(ctx context.Context) error {
	log.Logger.Info("Starting HostStatus controller")

	// Create and setup HostEndpoint informer
	hostEndpointInformer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				log.Logger.Debugf("Listing HostEndpoints for cluster agent: %s", c.config.ClusterAgentName)
				return c.client.HostEndpoints().List(context.Background(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				log.Logger.Debugf("Starting watch on HostEndpoints for cluster agent: %s", c.config.ClusterAgentName)
				return c.client.HostEndpoints().Watch(context.Background(), options)
			},
		},
		&bmcv1beta1.HostEndpoint{},
		0,
		cache.Indexers{},
	)

	_, err := hostEndpointInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: c.handleHostEndpointAdd,
		UpdateFunc: func(old, new interface{}) {
			c.handleHostEndpointAdd(new)
		},
		DeleteFunc: c.handleHostEndpointDelete,
	})
	if err != nil {
		log.Logger.Errorf("failed to add event handler for HostEndpoint: %v", err)
		return err
	}

	c.informer = hostEndpointInformer

	// Create and setup HostStatus informer
	hostStatusInformer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				log.Logger.Debugf("Listing HostStatus objects for cluster agent: %s", c.config.ClusterAgentName)
				return c.client.HostStatuses().List(context.Background(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				log.Logger.Debugf("Starting watch on HostStatus objects for cluster agent: %s", c.config.ClusterAgentName)
				return c.client.HostStatuses().Watch(context.Background(), options)
			},
		},
		&bmcv1beta1.HostStatus{},
		0,
		cache.Indexers{},
	)

	_, err = hostStatusInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleHostStatusAdd,
		UpdateFunc: c.handleHostStatusUpdate,
		DeleteFunc: c.handleHostStatusDelete,
	})
	if err != nil {
		log.Logger.Errorf("failed to add event handler for HostStatus: %v", err)
		return err
	}

	c.statusInformer = hostStatusInformer

	// Start informers first
	log.Logger.Debug("Starting informers...")
	go c.informer.Run(ctx.Done())
	go c.statusInformer.Run(ctx.Done())

	// Wait for the caches to be synced before starting workers
	log.Logger.Debug("Waiting for informer caches to sync")
	if !cache.WaitForCacheSync(ctx.Done(), c.informer.HasSynced, c.statusInformer.HasSynced) {
		return fmt.Errorf("failed to wait for caches to sync")
	}
	log.Logger.Debug("Informer caches synced successfully")

	c.wg.Add(2)
	log.Logger.Debug("Starting goroutines...")

	go func() {
		defer c.wg.Done()
		log.Logger.Debug("Starting DHCP event processor")
		c.processDHCPEvents(ctx)
		log.Logger.Debug("DHCP event processor exited")
	}()

	go func() {
		defer c.wg.Done()
		log.Logger.Debug("Starting HostEndpoint event processor")
		log.Logger.Debug("Waiting for items in workqueue...")
		for {
			select {
			case <-ctx.Done():
				log.Logger.Errorf("Context cancelled, stopping HostEndpoint event processor")
				return
			default:
				if ok := c.processNextWorkItem(); !ok {
					log.Logger.Errorf("processNextWorkItem returned false, stopping HostEndpoint event processor")
					return
				}
				log.Logger.Debug("Processed one item from workqueue, waiting for next...")
			}
		}
	}()

	go func() {
		c.UpdateHostStatusAtInterval()
	}()

	log.Logger.Debug("Both goroutines started")
	log.Logger.Info("HostStatus controller started successfully")
	return nil
}

func (c *hostStatusController) Stop() {
	log.Logger.Info("Stopping HostStatus controller")
	close(c.stopCh)
	c.workqueue.ShutDown()
	c.wg.Wait()
	log.Logger.Info("HostStatus controller stopped successfully")
}

func formatHostStatusName(agentName, ip string) string {
	return fmt.Sprintf("%s-%s", agentName, strings.ReplaceAll(ip, ".", "-"))
}

//----------------

// shouldRetry determines if an error should trigger a retry
func shouldRetry(err error) bool {
	return errors.IsConflict(err) || errors.IsServerTimeout(err) || errors.IsTooManyRequests(err)
}

func (c *hostStatusController) processNextWorkItem() bool {
	log.Logger.Debug("Trying to get next item from workqueue")
	obj, shutdown := c.workqueue.Get()
	if shutdown {
		log.Logger.Debug("Workqueue is shutdown")
		return false
	}
	defer c.workqueue.Done(obj)
	var err error
	var objName = "unknown"

	switch item := obj.(type) {
	case *bmcv1beta1.HostEndpoint:
		log.Logger.Debugf("Processing HostEndpoint from workqueue: %s", item.Name)
		objName = item.Name
		err = c.processHostEndpoint(item)
	case *bmcv1beta1.HostStatus:
		log.Logger.Debugf("Processing HostStatus from workqueue: %s", item.Name)
		objName = item.Name
		err = c.processHostStatus(item)
	default:
		log.Logger.Errorf("Unexpected type in workqueue: %s", reflect.TypeOf(obj))
		c.workqueue.Forget(obj)
		return true
	}

	if err == nil {
		c.workqueue.Forget(obj)
		return true
	}

	if shouldRetry(err) && c.workqueue.NumRequeues(obj) < maxRetries {
		log.Logger.Debugf("Error processing object %s (will retry): %v", objName, err)
		c.workqueue.AddRateLimited(obj)
		return true
	}

	// 如果重试次数超过限制，放弃处理
	log.Logger.Errorf("Dropping object after %d retries: %v", maxRetries, err)
	c.workqueue.Forget(obj)
	return true
}
