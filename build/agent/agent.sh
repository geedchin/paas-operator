#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

dir=$(cd $(dirname $0) && pwd)
cd $dir

sudo systemctl stop agent.service || true

sudo cp agent /usr/local/bin/agent
sudo cp ./agent.service /usr/lib/systemd/system/agent.service

sudo systemctl daemon-reload
sudo systemctl enable agent.service
sudo systemctl restart agent.service
