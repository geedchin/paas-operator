#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

function kube::build::build_image() {
  OUTPUT_DIR=_output
  ROOT_PATH=$(dirname "${BASH_SOURCE[0]}")/..

  cd ${ROOT_PATH}/cmd/agent/ && go build . && cd -
  cd ${ROOT_PATH}/cmd/apiserver/ && go build . && cd -

  cp ${ROOT_PATH}/cmd/agent/agent ${ROOT_PATH}/build/agent/
  cp ${ROOT_PATH}/cmd/apiserver/apiserver ${ROOT_PATH}/build/apiserver/docker/

  cd ${ROOT_PATH}/build/ && tar -czvf agent.tar.gz agent && cd -

  cp ${ROOT_PATH}/build/agent.tar.gz ${ROOT_PATH}/build/apiserver/docker/
  cd ${ROOT_PATH}/build/apiserver/docker/ && docker build -t apiserver:v1 . && cd -

  mkdir -p ${ROOT_PATH}/build/${OUTPUT_DIR}
  cd ${ROOT_PATH}/build/${OUTPUT_DIR}/
  docker save apiserver:v1 -o apiserver_v1.tar
  cd -

  # mv ${ROOT_PATH}/build/agent.tar.gz ${ROOT_PATH}/build/${OUTPUT_DIR}/
}