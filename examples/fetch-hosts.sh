#!/bin/bash

# Build a minimal Linux kernel without modules.

set -e -x -o pipefail

# Example by Alex Ellis

mkdir -p uploads
sudo cp /etc/hostname uploads/
sudo cp /etc/hosts uploads/
sudo cp /etc/resolv.conf uploads/

