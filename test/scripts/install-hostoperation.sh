#!/bin/bash

hostStatusName=$1

echo "hostStatusName: $hostName"

[ -n "${hostStatusName}" ] || {
    echo "HostStatusName is required"
    kubectl get hoststatus
    exit 1
}

kubectl get hoststatus ${hostStatusName} &>/dev/null || {
    echo "HostEndpoint ${hostStatusName} not found"
    kubectl get hoststatus
    exit 1
}

# 创建测试用的 HostOperation 实例
cat <<EOF | kubectl apply -f -
apiVersion: bmc.spidernet.io/v1beta1
kind: HostOperation
metadata:
  name: testop
spec:
  action: powerOff
  hostStatusName: ${hostStatusName}
EOF

echo "HostOperation for ${hostStatusName} created"

