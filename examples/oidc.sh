#!/bin/bash

# Warning: it's recommend to only run this with the --private (repo) flag

env

OIDC_TOKEN=$(curl -sLS "${ACTIONS_ID_TOKEN_REQUEST_URL}&audience=https://fed-gw.exit.o6s.io" -H "User-Agent: actions/oidc-client" -H "Authorization: Bearer $ACTIONS_ID_TOKEN_REQUEST_TOKEN")
JWT=$(echo $OIDC_TOKEN | jq -j '.value')

jq -R 'split(".") | .[1] | @base64d | fromjson' <<< "$JWT"

# Post the JWT to the printer function to visualise it in the logs
# curl -sLSi ${OPENFAAS_URL}/function/printer -H "Authorization: Bearer $JWT"
