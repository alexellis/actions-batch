#!/bin/bash

set -e -x -o pipefail

# Example by Alex Ellis

# Build and then export a Docker image to a tar file
# The exported file can then be imported into your local library via:

# docker load -i curl.tar

mkdir -p uploads

cat <<'EOF' >Dockerfile
FROM alpine:latest

RUN apk --no-cache add curl

ENTRYPOINT ["curl"]
EOF

docker build -t curl:latest .

# export the image to a tar file

docker save curl:latest >uploads/curl.tar
