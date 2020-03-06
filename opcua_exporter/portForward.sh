#!/bin/bash
# Tunnel into a specific EC2 server to get to San Diego
# Expose the ignition OPCUA endpoint on localhost:4096
ssh -L 4096:172.19.12.46:4096 ubuntu@54.191.199.94
