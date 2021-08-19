#!/usr/bin/env bash

if [ -z "$1" ]
then
  echo "Usage: ./generate-test-incentive-data.sh genesis.json"
  exit 1
fi

jq < "$1" -c '{"incentive": .app_state.incentive, "cdp": .app_state.cdp }' > testdata/kava-7-test-incentive-state.json
