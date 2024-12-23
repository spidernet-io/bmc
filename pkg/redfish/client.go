package redfish

import (
	"github.com/stmcginnis/gofish"
)

// Health 实现健康检查方法
func (c *redfishClient) Health() bool {
	// 创建 gofish 客户端
	client, err := gofish.Connect(c.config)
	if err != nil {
		return false
	}
	defer client.Logout()
	return true
}
