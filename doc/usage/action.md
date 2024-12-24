# HostOperation 操作指南

本文档介绍了如何使用 HostOperation CRD 来管理物理机的电源状态。

## 支持的操作类型

HostOperation CRD 支持以下四种操作类型：

| Action | 描述 | 使用场景 |
|--------|------|----------|
| powerOff | 关闭物理机电源 | 需要维护物理机或节能时 |
| powerOn | 开启物理机电源 | 物理机关机状态下，需要启动时 |
| reboot | 重启物理机 | 物理机系统异常需要重启，或更新系统后需要重启时 |
| pxeReboot | 通过 PXE 方式重启物理机，它是 once PXE 重启效果 | 需要重新安装操作系统，或进行 PXE 引导时 |

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

关闭物理机电源：
```yaml
apiVersion: bmc.spidernet.io/v1beta1
kind: HostOperation
metadata:
  name: host-1-poweroff
spec:
  action: powerOff
  hostStatusName: host-1
```

开启物理机电源：
```yaml
apiVersion: bmc.spidernet.io/v1beta1
kind: HostOperation
metadata:
  name: host-1-poweron
spec:
  action: powerOn
  hostStatusName: host-1
```

重启物理机：
```yaml
apiVersion: bmc.spidernet.io/v1beta1
kind: HostOperation
metadata:
  name: host-1-reboot
spec:
  action: reboot
  hostStatusName: host-1
```

PXE 重启物理机：
```yaml
apiVersion: bmc.spidernet.io/v1beta1
kind: HostOperation
metadata:
  name: host-1-pxereboot
spec:
  action: pxeReboot
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
  name: host-1-poweroff
spec:
  action: powerOff
  hostStatusName: host-1
EOF
```

3. 查看操作状态：
```bash
# 查看所有操作
kubectl get hostoperation

# 查看特定操作的详细信息
kubectl get hostoperation host-1-poweroff -o yaml

# 监控操作状态变化
kubectl get hostoperation host-1-poweroff -w
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
5. 对于 PXE 重启，它是实现 once 重启，即重启后。需要管理员在带内网络内手动部署 PXE 服务，本组件并不自动部署 PXE 服务

### 最佳实践

1. 在执行关键操作前，建议：
   - 备份重要数据
   - 选择合适的维护时间窗口
   - 通知相关人员

2. 操作命名建议：
   - 使用小写字母
   - 使用有意义的前缀
   - 包含操作类型信息
   - 示例：`host-1-poweroff`、`host-2-pxereboot`

3. 批量操作时：
   - 建议分批执行
   - 每批之间留有观察时间
   - 确保有回滚方案
