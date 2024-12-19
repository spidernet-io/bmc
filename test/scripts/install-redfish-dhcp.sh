#!/bin/bash

set -x
set -o errexit
set -o pipefail
set -o nounset

CURRENT_FILENAME=$( basename $0 )
CURRENT_DIR_PATH=$(cd $(dirname $0); pwd)
PROJECT_ROOT_PATH=$(cd ${CURRENT_DIR_PATH}/../..; pwd)

TOOLS_IMAGE_REF=ghcr.io/spidernet-io/bmc-tools:latest

# try to load tools image
docker inspect ${TOOLS_IMAGE_REF} &>/dev/null || \
    docker pull ${TOOLS_IMAGE_REF} || \
    ( cd ${PROJECT_ROOT_PATH} && make build-tools-image )

IMAGES=$( helm template redfish ${CURRENT_DIR_PATH}/../redfishchart | grep "image:"  | awk '{print $2}' | sort | tr -d '"' | uniq )
echo "IMAGES"
echo "${IMAGES}"
for IMAGE in $IMAGES; do
    echo "loading $IMAGE"
    docker inspect $IMAGE &>/dev/null || docker pull $IMAGE 
    kind load docker-image $IMAGE --name ${E2E_CLUSTER_NAME}
done

echo "install redfish"
helm uninstall dhcp-redfish -n  redfish || true 
helm install dhcp-redfish ${CURRENT_DIR_PATH}/../redfishchart \
  --wait \
  --debug \
  --namespace redfish \
  --create-namespace \
  --set replicaCount=2  \
  --set networkInterface=net1  \
  --set underlayMultusCNI="${UNDERLAY_CNI}"
