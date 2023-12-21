#!/bin/bash

# Example by Alex Ellis

git clone --depth=1 https://github.com/openfaas/faas-netes
cd faas-netes

SERVER=ttl.sh make build-buildx

# Run the end to end tests

./contrib/get_tools.sh
./contrib/lint_chart.sh
./contrib/create_cluster.sh
OPERATOR=0 ./contrib/deploy.sh
OPERATOR=0 ./contrib/run_function.sh
./contrib/stop_dev.sh
