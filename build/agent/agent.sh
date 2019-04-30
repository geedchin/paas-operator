#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

sudo cp agent /usr/local/bin/agent
sudo cp ./agent.service /usr/lib/systemd/system/agent.service

sudo systemctl daemon-reload
sudo systemctl enable agent.service
sudo systemctl start agent.service
