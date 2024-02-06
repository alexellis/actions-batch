#!/bin/bash

# Example by BadgerOps
# Sysad script that checked who was logged in & what program they were running

set -e -x -o pipefail

echo "User Login and Process Monitoring Script"
echo "----------------------------------------"

echo "Currently logged in users:"
who

echo
echo "Processes by user:"
echo "----------------------------------------"

for user in $(who | awk '{print $1}' | sort | uniq)
do
    echo "User: $user"
    echo "Processes:"
    ps -u $user --no-headers
    echo "----------------------------------------"
done

echo "Script completed."
