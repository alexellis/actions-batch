#!/bin/bash

# Example by Alex Ellis

set -e -x -o pipefail

sudo apt update -qqy && \
    sudo apt install -qqy \
        cowsay \
        fortune \
        --no-install-recommends

fortune | cowsay
