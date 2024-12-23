package data

import (
	"sync"

	"github.com/spidernet-io/bmc/pkg/k8s/apis/bmc.spidernet.io/v1beta1"
)

// HostConnectCon 定义主机数据结构
type HostConnectCon struct {
	Info     *v1beta1.BasicInfo
	Username string
	Password string
}

// HostCache 定义主机缓存结构
type HostCache struct {
	lock sync.RWMutex
	data map[string]HostConnectCon
}

var HostCacheDatabase *HostCache

func init() {
	HostCacheDatabase = &HostCache{
		data: make(map[string]HostConnectCon),
	}
}

// Add 添加或更新缓存中的主机数据
func (c *HostCache) Add(name string, data HostConnectCon) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.data[name] = data
}

// Delete 从缓存中删除指定主机数据
func (c *HostCache) Delete(name string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	delete(c.data, name)
}

// Get 获取指定主机的数据
func (c *HostCache) Get(name string) (HostConnectCon, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	data, exists := c.data[name]
	return data, exists
}
