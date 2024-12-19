#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail
set -x

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
PROJECT_ROOT=$(git rev-parse --show-toplevel)
CODEGEN_PKG=/tmp/code-generator
MODULE_NAME=$(cat ${PROJECT_ROOT}/go.mod | grep -e "module[[:space:]][^[:space:]]*" | awk '{print $2}')

# API and output package paths
APIS_PKG="pkg/k8s/apis"
OUTPUT_PKG="pkg/k8s/client"

# Setup temporary directory for code generation
TMP_DIR="${PROJECT_ROOT}/output/codeGen"
LICENSE_FILE="${TMP_DIR}/boilerplate.go.txt"

# Clean and create temporary directory
rm -rf ${TMP_DIR}
mkdir -p ${TMP_DIR}

# Create license header file
touch ${LICENSE_FILE}
while read -r line || [[ -n ${line} ]]
do
    echo "// ${line}" >>${LICENSE_FILE}
done < "${PROJECT_ROOT}/tools/copyright-header.txt"

# Clean existing output
rm -rf ${OUTPUT_PKG} || true

# Source the code generator script
source "${CODEGEN_PKG}/kube_codegen.sh"

# Generate client API
echo "generate client api"
kube::codegen::gen_helpers \
  --input-pkg-root "${MODULE_NAME}/${APIS_PKG}" \
  --output-pkg-root "${MODULE_NAME}/${OUTPUT_PKG}" \
  --boilerplate "${LICENSE_FILE}" \
  --output-base "${PROJECT_ROOT}"

kube::codegen::gen_client \
  --with-watch \
  --with-applyconfig \
  --input-pkg-root "${MODULE_NAME}/${APIS_PKG}" \
  --output-pkg-root "${MODULE_NAME}/${OUTPUT_PKG}" \
  --boilerplate "${LICENSE_FILE}" \
  --output-base "${PROJECT_ROOT}"

# Cleanup
rm -rf ${TMP_DIR}
exit 0
