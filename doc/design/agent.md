# agent

## yaml

- 需要 controller 的 生成 agent deployment 代码中，配置 健康检查的端口和配置， agent 通过 启动的命令参数，在指定的端口上 进行 启动 http server 提供健康检查


## yaml

- 需要 controller 的 生成 agent deployment 代码中，配置 健康检查的端口和配置， agent 通过 启动的命令参数，在指定的端口上 进行 启动 http server 提供健康检查


## 启动

agent 启动的初始代码中，获取 环境变量 CLUSTERAGENT_NAME 的值， 通过该来 作为 name ，去获取 相应的 clusteragent 实例的 yaml 定义， 使用其 spec.endpoint 和 spec.feature 中的 内容 生成 一组 struct config 配置，作为后续工作的基础

type agentConfig struct {
  ClusterAgentName string
  Endpoint         *EndpointConfig
  Feature          *FeatureConfig      
}

获取 config 后，需要对其进行 如下 校验
（1） spec.endpoint.https 如果该值为 true，需要 spec.endpoint.secretName 和 spec.endpoint.secretNamespace 必须有值， 否则 程序报错退出
（2） spec.endpoint.secretName 代表 secret name ， spec.endpoint.secretNamespace 代表 secret namespace， 如果他们有值，需要确认该 secret 是否存在，其中 要求 存在 tls 认证的 key 和 cert ，但不 要求一定有 ca   。  否则 程序报错退出
（3）spec.feature. dhcpServerInterface 必须有值，且确认该值所代表的网卡  是否存在于 网络接口中，且网卡 up 状态， 否则 程序报错退出


## dhcp server 模块

在 pkg/dhcpserver 下 实现一个 接口模块，使用 interface 对外暴露，  它应该具备如下方法，并实现它们

1 启动 dhcp server 接口
    它基于 dhcpd 二进制来启动 dhcp server 服务

    它基于 AgentConfig.objSpec.Feature.DhcpServerConfig 中的参数 工作
    模块的参数，工作的网卡名参数（必备），暴露 dhcp 分配 ip 的子网 参数（必备），可分配 ip 参数（必备） , 分配给 client 端的 子网网关 ip 参数（必备）。 这些参数 传递给 dhcpd 进行工作
    
    AgentConfig.objSpec.Feature.DhcpServerConfig.selfIp 如果有值， 那么 把 网卡 AgentConfig.objSpec.Feature.DhcpServerConfig.DhcpServerInterface  名参数所代表的 网卡上的 IP 地址 去除，然后 用 网卡 ip 参数 代替 生效

    模块一直监控 dhcpd 的运行，当 它 故障时，能够再次 尝试 拉起 它 

    模块一直监控 dhcpd 的运行，当 分配 或者 释放 一个 ip 时，有 日志 输出，该行日志中 包括 可分配 ip 的 总量和剩余量


2 获取 client 信息接口 
   接口输出 获取 dhcpd 分配出的 所有 client 的 ip 地址 和 对应的 mac 的列表

3 获取 ip 用量统计信息
   接口输出 当前可分配 ip 的 总量和剩余量 等

4 关闭 dhcp server 接口， 停止 dhcpd 服务


agent 的  main 函数 主框架中， 根据 自身的AgentConfig.objSpec.Feature.enableDhcpServer 配置，来决定 是否要调用 pkg/dhcpserver 模块中的接口，来启动 dhcp server


需要 考虑 dhcpd 的 /var/lib/dhcp/dhcpd.leases 持久化问题，这样，在 agent 重启后，dhcpd 的数据可以持久化，不会丢失
这样，在 helm 中的 values.yaml 中， 需要支持 二选一的 方式 来 进行持久化 
（1）在调试环境中，使用宿主机的 本地挂载， 把 宿主机的 /var/lib/dhcp/ 目录 挂载给 agent pod  的 /var/lib/dhcp/ 
（2）在生产环境中，可使用 pvc 来存储 dhcpd 的数据，把 pvc 挂载给 agent pod 的 /var/lib/dhcp/
    NewDhcpServer 函数中，需要额外传入  clusterAgentName， 该变量用于生成 dhcp server 的 lease  文件名 /var/lib/dhcp/${clusterAgentName}-dhcpd.leases ， 给 dhcp server 使用， 这样，即使在 本地 磁盘持久阿虎是，能实现 文件的不冲突 
    
    

## crd hostEndpoint


