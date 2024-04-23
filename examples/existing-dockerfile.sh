#!/bin/bash

set -e -x -o pipefail

# Example by Moulick Aggarwal

# Build a existing Dockerfile and then export a Docker image to a tar file
# The exported file can then be imported into your local library via:

# docker load -i curl.tar

mkdir -p uploads

docker build -t curl:latest .

# export the image to a tar file

docker save curl:latest > uploads/curl.tar
