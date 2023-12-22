#!/bin/bash

# Build a minimal Linux kernel without modules.

set -e -x -o pipefail

# Example by Alex Ellis

export BRANCH=${BRANCH:-"v6.0"}

echo "Installing build dependencies"

time sudo apt update -qqqy && \
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

echo "Cloning Linux Kernel branch $BRANCH"
time git clone https://github.com/torvalds/linux.git linux.git --depth=1 --branch $BRANCH && \
    cd linux.git && \
    scripts/config --disable SYSTEM_TRUSTED_KEYS && \
    scripts/config --disable SYSTEM_REVOCATION_KEYS && \
    make oldconfig

if [ "$(uname -m)" = "aarch64" ]; then
    echo "Building for ARM64"
    time make Image -j$(nproc) && mkdir -p uploads

    # Save the resulting Kernel binary so that it's downloaded to the user's computer
    cp ./arch/arm64/boot/Image ./uploads/
else
    echo "Building for x86_64"
    make -j$(nproc) vmlinux && mkdir -p uploads

    # Save the resulting Kernel binary so that it's downloaded to the user's computer
    cp vmlinux ./uploads/
fi

echo "Build complete, size of Kernel:"

du -h ./uploads/
