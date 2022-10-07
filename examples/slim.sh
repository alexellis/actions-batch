#!/bin/bash

set -e -x -o pipefail

# Example from: https://github.com/docker-slim/docker-slim

curl -S -L https://downloads.dockerslim.com/releases/1.38.0/dist_linux.tar.gz -o /tmp/dist_linux.tar.gz && \
    sudo tar -xvf /tmp/dist_linux.tar.gz --strip-components=1 -C /usr/local/bin/ && \
    rm /tmp/dist_linux.tar.gz

docker pull archlinux:latest

docker-slim build --target archlinux:latest --tag archlinux:curl \
     --http-probe=false --exec "curl checkip.amazonaws.com"

docker run archlinux:curl curl checkip.amazonaws.com

docker images
