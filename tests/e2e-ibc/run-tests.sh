#!/usr/bin/env bash
set -euo pipefail

go_bin=$1
listing=$("$go_bin" test -list '^Test' .)
suites=$(grep '^Test' <<< "$listing")

while read -r suite; do
    # interchaintest probes loopback ports, then publishes them on all interfaces.
    # On Linux, RPC connections from the previous suite's Docker bridge can
    # still occupy those ports in TIME_WAIT for up to a minute.
    if command -v ss >/dev/null; then
        deadline=$((SECONDS + 65))
        while (( SECONDS < deadline )) && [[ -n $(ss -Htan state time-wait '( dport = :26657 )') ]]; do
            sleep 2
        done
    fi
    echo "Running $suite"
    "$go_bin" test -failfast -timeout 10m -count=1 -run "^${suite}$" .
done <<< "$suites"
