package redfish

import (
	"fmt"

	"github.com/spidernet-io/bmc/pkg/agent/hoststatus/data"
	"github.com/spidernet-io/bmc/pkg/log"
	"github.com/stmcginnis/gofish"
	"go.uber.org/zap"
)

// Client 定义了 Redfish 客户端接口
type RefishClient interface {
	// Health 检查 Redfish 服务的健康状态
	Health() bool
	Reboot(BootCmd) error
	GetInfo() error
}

// redfishClient 实现了 Client 接口
type redfishClient struct {
	config gofish.ClientConfig
	logger *zap.SugaredLogger
}

var _ RefishClient = (*redfishClient)(nil)

// NewClient 创建一个新的 Redfish 客户端
func NewClient(hostCon data.HostConnectCon) RefishClient {

	url := buildEndpoint(hostCon)
	config := gofish.ClientConfig{
		Endpoint: url,
		Username: hostCon.Username,
		Password: hostCon.Password,
		Insecure: true,
	}
	return &redfishClient{
		config: config,
		logger: log.Logger.Named("redfish").With(
			zap.String("endpoint", url),
		),
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
