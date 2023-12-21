#!/bin/bash

# Example by Alex Ellis

# https://go.dev/blog/govulncheck

set -e -x -o pipefail

curl -sLS https://get.arkade.dev | sudo sh
sudo arkade system install go --progress=false

export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin

go install golang.org/x/vuln/cmd/govulncheck@latest

git clone --depth=1 https://github.com/openfaas/faas
cd faas/gateway

govulncheck .

