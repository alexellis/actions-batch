#!/bin/bash

# Build a minimal Linux kernel without modules.

set -e -x -o pipefail

# Example by Alex Ellis

echo "Installing build dependencies"

sudo apt update -qqqy && \
sudo apt-get install -qqqy \
    git \
    build-essential \
    fakeroot \
    libncurses5-dev \
    libssl-dev \
    ccache \
    bison \
    flex \
    libelf-dev \
    dwarves \
    bc

time git clone https://github.com/torvalds/linux.git linux.git --depth=1 --branch v6.0

cd linux.git 
scripts/config --disable SYSTEM_TRUSTED_KEYS
scripts/config --disable SYSTEM_REVOCATION_KEYS
make oldconfig

time make vmlinux -j$(nproc)
du -h ./vmlinux


# Save the resulting Kernel binary so that it's downloaded to the user's computer
cp vmlinux ./uploads/
