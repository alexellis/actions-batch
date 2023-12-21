#!/bin/bash

# This script builds the arkade binary for all platforms.
# Then they're downloaded to your machine directly.

# Arkade is a trivial program to build, but it shows that you
# don't need Go to be installed on your system.

set -e -x -o pipefail

# Example by Alex Ellis

curl -sLS https://get.arkade.dev | sudo sh
sudo arkade system install go --progress false

export PATH=$PATH:$HOME/.arkade/bin:$HOME/go/bin:/usr/local/go/bin

mkdir -p uploads

git clone --depth=1 https://github.com/alexellis/arkade
cd arkade

make dist && \
    mv ./bin/* ../uploads/
