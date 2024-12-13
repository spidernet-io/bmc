# 设计

写一个 kubernetes 下的 组件

## controller 组件

1.它能够通过 helm 进行 安装，必要的安装参数，请基于 helm values 进行暴露
安装后，默认安装 一个 2 pod 的 deployment 的 controller 组件

- helm 的 chart 代码 放置在工程的 chart 目录下

- helm中的 deployment 的 镜像名为 spidernet-io/bmc/controller

- controller pod 中的 进程 golang 代码放在 cmd/controller 目录下，所有的代码，要求有详细的 debug 级别日志，必要的 log 级别和 error 级别的日志，有必要的 代码注释

- chart 的版本为工程根目录下的 VERSION 中的值

- deployment yaml 中 有 pod 健康检查的设置 ， 访问 8000 端口

2.helm  安装 了如下 crd

apiVersion: bmc.io/v1beta1
kind: clusterAgent
metadata:
  name: test
spec:
  interface: “kube-system/macvlan-pod-network”
  clusterName: "cluster1"
status:
  ready: true

所有的 crd 定义的 golang 代码， 放置在 pkg/apis 的 相关目录下


3. 当用户创建了 crd clusterAgent 的实例后，controller 的 golang 进行在 kubernetes 中 创建出一个 单 pod 的 deployment 实例 ， 它名为 agent 组件

- agent 组件 的  deployment 的 实例名， 为 crd clusterAgent 的 metadata.name

- agent 组件 的  yaml 中，在 8000 端口上的健康检查配置

- agent 组件 的  yaml 中，配置一些与该实例对应的 必要 label 和 annotation 

- 基于 crd clusterAgent 中的 spec.interface 的值，结合 “k8s.v1.cni.cncf.io/networks” 的 key 名， 该 agent 组件 的 deployment 的 annotation 中带有如下
k8s.v1.cni.cncf.io/networks: kube-system/macvlan-pod-network

- 基于 crd clusterAgent 中的 spec.clusterName 的值， 该 agent 组件 的 deployment 注入 环境变量，环境变量的 key  为 ClusterName， 环境变量的值为 spec.clusterName

- controller 的进程能够 监控 每个  crd clusterAgent 对应的 deployment 实例的状态，如果它的所有 pod 是 running 的，那么就 标记 CRD clusterAgent 中的 status.ready=true， 否则 status.ready=false

4. controller 代码 具备以下编程要点

- 能监听在的端口 8000 上，提供必要的 http 接口，用于进行 pod 的健康检查

## agent 组件

agent pod 中的 进程 golang 代码放在 cmd/agent 目录下， 所有的代码，要求有详细的 debug 级别日志，必要的 log 级别和 error 级别的日志
，有必要的 代码注释

- agent 的 server 进程，默认一直在 sleep 

- agent 组件 的  yaml 中，在 8000 端口上的健康检查配置


## 镜像 

构建镜像，遵循如下规则

- controller pod 的 镜像 构建 docker 位于 image/controller 目录下， 它基于 ubuntu 基础镜像构建，在 dockerfile 中，把整个工程 copy 进行进行，使用工程根目录下 makefile 中的能力，把 cmd/controller 目录下的 golang 代码进行编译，最终打包进入 镜像，作为镜像的启动程序入口 

- agent pod 的 镜像 构建 docker 位于 image/agent 目录下，它基于 ubuntu 基础镜像构建，在 dockerfile 中，把整个工程 copy 进行进行，使用工程根目录下 makefile 中的能力，把 cmd/agent 目录下的 golang 代码进行编译，最终打包进入 镜像，作为镜像的启动程序入口 


## makefile

工程根目录下，有个 makefile 文件，它是一个 makefile 的工程，可以使用 make 命令进行编译， 它要支持：
- 基于  image/controller/Dockerile， 能够生成 controller 的 docker 镜像，镜像名为  spidernet-io/bmc/controller ，镜像的 tag 为工程根目录下的 VERSION 中的值

- 基于  image/agent/Dockerile， 能够生成 server 的 docker 镜像，镜像名为  spidernet-io/bmc/server，镜像的 tag 为工程根目录下的 VERSION 中的值
 
- 生成 helm chart 

- 关于 makefile 的用法，支持 使用 make usage 进行输出显示

## 整个工程的编码

- 整个golang 的工程为 github.com/spidernet-io/bmc ， 所以，所有功能内的相关代码引用，请遵循该规范

- 工程中所有的 golang、shell代码文件中，都使用英文，包括代码注释和日志

## 文档

在 根目录的 README.md 中, 描述一下内容：
- 整个工程的功能描述
- kubernetes 基本安装方法

在工程的所有 golang 代码目录下，创建必要的 readme.md 文件，对相关目录下的代码做简要的设计说明

