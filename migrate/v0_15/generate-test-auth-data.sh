#!/usr/bin/env bash

# Usage: ./generate-test-auth-data.sh genesis.json
jq < "$1" '.app_state.auth | { params: .params, accounts: [.accounts[] | select((.value.coins | length) > 0)] }' > testdata/kava-7-test-auth-state.json
