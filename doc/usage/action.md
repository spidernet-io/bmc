# HostOperation 操作指南

本文档介绍了如何使用 HostOperation CRD 来管理物理机的电源状态。

## 支持的操作类型

HostOperation CRD 支持以下操作类型：

| Action | 描述 | 使用场景 |
|--------|------|----------|
| On | 正常开机 | 物理机关机状态下，需要启动时 |
| ForceOn | 强制开机 | 物理机可能处于异常关机状态时 |
| ForceOff | 强制关机，强制操作会立即执行，可能导致数据丢失 | 物理机无法正常关机时 |
| GracefulShutdown | 优雅关机，优雅操作会等待操作系统完成清理工作 | 正常关闭物理机，等待操作系统完成清理 |
| ForceRestart | 强制重启，强制操作会立即执行，可能导致数据丢失 | 物理机系统无响应需要强制重启时 |
| GracefulRestart | 优雅重启，优雅操作会等待操作系统完成清理工作 | 正常重启物理机，等待操作系统完成清理 |
| PxeReboot | PXE 重启，PXE 重启是实现 once 重启，即重启后。需要管理员在带内网络内手动部署 PXE 服务，本组件并不自动部署 PXE 服务 | 需要通过 PXE 引导安装系统时 |

## 操作流程

### 前提条件

1. 确保物理机已经被正确注册到系统中（存在对应的 HostStatus CR）
2. 确保物理机状态健康（HostStatus 的 status.healthy 为 true）
3. 确保有足够的权限执行这些操作

### 操作步骤

1. 查看当前可操作的物理机列表：
```bash
kubectl get hoststatus
```

2. 创建 HostOperation：

以下是不同操作类型的 YAML 示例：

正常开机：
```yaml
apiVersion: bmc.spidernet.io/v1beta1
kind: HostOperation
metadata:
  name: host-1-on
spec:
  action: On
  hostStatusName: host-1
```

强制关机：
```yaml
apiVersion: bmc.spidernet.io/v1beta1
kind: HostOperation
metadata:
  name: host-1-forceoff
spec:
  action: ForceOff
  hostStatusName: host-1
```

优雅重启：
```yaml
apiVersion: bmc.spidernet.io/v1beta1
kind: HostOperation
metadata:
  name: host-1-gracefulrestart
spec:
  action: GracefulRestart
  hostStatusName: host-1
```

PXE 重启：
```yaml
apiVersion: bmc.spidernet.io/v1beta1
kind: HostOperation
metadata:
  name: host-1-pxereboot
spec:
  action: PxeReboot
  hostStatusName: host-1
```

创建操作：
```bash
# 将上述 YAML 保存为文件（如 hostop.yaml）后执行：
kubectl apply -f hostop.yaml

# 或直接使用 kubectl create 命令：
cat <<EOF | kubectl create -f -
apiVersion: bmc.spidernet.io/v1beta1
kind: HostOperation
metadata:
  name: host-1-gracefulshutdown
spec:
  action: GracefulShutdown
  hostStatusName: host-1
EOF
```

3. 查看操作状态：
```bash
# 查看所有操作
kubectl get hostoperation

# 查看特定操作的详细信息
kubectl get hostoperation host-1-gracefulshutdown -o yaml

# 监控操作状态变化
kubectl get hostoperation host-1-gracefulshutdown -w
```

### 状态说明

操作状态可以通过 `status.status` 字段查看：

| 状态 | 描述 |
|------|------|
| pending | 操作正在执行中 |
| success | 操作执行成功 |
| failed | 操作执行失败 |

### 注意事项

1. 每个操作都是一次性的，不支持更新已创建的 HostOperation
2. 建议使用有意义的命名方式，如 `<hostStatusName>-<action>` 格式
3. 在执行操作前，请确保：
   - 物理机处于健康状态
   - 了解操作可能带来的影响
   - 有足够的权限执行操作
4. 如果操作失败：
   - 查看 `status.message` 了解失败原因
   - 检查物理机状态和网络连接
   - 确保 BMC 接口可访问

### 最佳实践

1. 在执行关键操作前，建议：
   - 备份重要数据
   - 选择合适的维护时间窗口
   - 通知相关人员

2. 操作命名建议：
   - 使用小写字母
   - 使用有意义的前缀
   - 包含操作类型信息
   - 示例：`host-1-gracefulshutdown`、`host-2-pxereboot`

3. 批量操作时：
   - 建议分批执行
   - 每批之间留有观察时间
   - 确保有回滚方案

4. 操作选择建议：
   - 优先使用优雅操作（GracefulShutdown/GracefulRestart）
   - 仅在必要时使用强制操作（ForceOff/ForceRestart）
   - 确保了解每种操作的影响
