package redfish

import (
	"fmt"

	"github.com/spidernet-io/bmc/pkg/agent/hoststatus/data"
	"github.com/stmcginnis/gofish"
)

// Client 定义了 Redfish 客户端接口
type Client interface {
	// Health 检查 Redfish 服务的健康状态
	Health() bool
}

// NewClient 创建一个新的 Redfish 客户端
func NewClient(hostCon data.HostConnectCon) Client {
	return newRedfishClient(hostCon)
}

// redfishClient 实现了 Client 接口
type redfishClient struct {
	config gofish.ClientConfig
}

// newRedfishClient 创建一个新的 Redfish 客户端
func newRedfishClient(hostCon data.HostConnectCon) Client {
	config := gofish.ClientConfig{
		Endpoint: buildEndpoint(hostCon),
		Username: hostCon.Username,
		Password: hostCon.Password,
		Insecure: true,
	}
	return &redfishClient{
		config: config,
	}
}

// buildEndpoint 根据 HostConnectCon 构建 Redfish 服务的端点 URL
func buildEndpoint(hostCon data.HostConnectCon) string {
	protocol := "http"
	if hostCon.Info.Https {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s:%d", protocol, hostCon.Info.IpAddr, hostCon.Info.Port)
}
