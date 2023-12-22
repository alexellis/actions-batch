#!/bin/bash

# Run the grype container image scanner against a list of images

set -e -x -o pipefail

# Example by Alex Ellis

if [ ! command -v arkade &> /dev/null ]
then
    curl -sLS https://get.arkade.dev | sudo sh
    export PATH=$PATH:$HOME/.arkade/bin
fi

arkade get grype --quiet

# Bash array of ghcr.io images for openfaas
IMAGES=("ghcr.io/openfaas/gateway:$(crane ls ghcr.io/openfaas/gateway|tail -n 1)" "ghcr.io/openfaas/faas-netes:$(crane ls ghcr.io/openfaas/faas-netes|tail -n 1)")

for image in "${IMAGES[@]}"
do
    echo "Running grype against: $image"
    grype $image --quiet
done
