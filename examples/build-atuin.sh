#!/bin/bash

# This script builds the atuin binary 
# Example by Ellie Huxtable

set -e -x -o pipefail

curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y

source $HOME/.cargo/env

mkdir -p uploads

git clone --depth=1 https://github.com/atuinsh/atuin
cd atuin

cargo build --release

cp target/release/atuin ../uploads/atuin
