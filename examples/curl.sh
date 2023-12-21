#!/bin/bash

set -e -x -o pipefail

# Example by Alex Ellis

curl -s -X POST https://hookb.in/XklqazrWNBFDkmwD3VLb -d "Logged in as: $(whoami)"
