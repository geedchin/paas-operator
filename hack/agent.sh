#!/bin/sh

sudo cp agent /usr/local/bin/agent
sudo cp ./agent.service /usr/lib/systemd/system/agent.service
sudo systemctl enable agent.service
sudo systemctl daemon-reload
sudo systemctl start agent.service
