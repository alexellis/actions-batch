#!/bin/bash

set -e -x -o pipefail

# Example by Alex Ellis

# Create a folder called .secrets
#
# Add a file called .secrets/openfaas-gateway-password with your admin user
# Then create another file called .secrets/openfaas-url with the URL of your OpenFaaS gateway
# Use a personal inlets subscription to expose it or a public cloud load blancer
#
# This will be created as a repo-level secret named OPENFAAS_GATEWAY_PASSWORD

curl -sLS https://get.arkade.dev | sudo sh

arkade get faas-cli --quiet
sudo mv $HOME/.arkade/bin/faas-cli /usr/local/bin/
sudo chmod +x /usr/local/bin/faas-cli 

echo "${OPENFAAS_GATEWAY_PASSWORD}" | faas-cli login -g "${OPENFAAS_URL}" -u admin --password-stdin

# List some functions
faas-cli list

# Deploy a function to show this worked and update the "com.github.sha" annotation
faas-cli store deploy env --name env-actions-batch --annotation com.github.sha=${GITHUB_SHA}

sleep 2

# Invoke the function
faas-cli invoke env-actions-batch <<< ""

