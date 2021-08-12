#!/usr/bin/env bash

if [ -z "$1" ]
then
  echo "Usage: ./generate-test-auth-data.sh genesis.json"
  exit 1
fi

jq < "$1" -c '.app_state.auth | { params: .params, accounts: [.accounts[] | select((.value.coins | length) > 0)] }' > testdata/kava-7-test-auth-state.json
