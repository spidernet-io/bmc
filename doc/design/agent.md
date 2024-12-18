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


## crd hostEndpoint


