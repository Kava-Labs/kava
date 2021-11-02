<!--
title: swaps
-->
## kvcli query bep3 swaps

query atomic swaps with optional filters

### Synopsis

Query for all paginated atomic swaps that match optional filters:
Example:
$ kvcli q bep3 swaps --involve=kava1l0xsq2z7gqd7yly0g40y5836g0appumark77ny
$ kvcli q bep3 swaps --expiration=280
$ kvcli q bep3 swaps --status=(Open|Completed|Expired)
$ kvcli q bep3 swaps --direction=(Incoming|Outgoing)
$ kvcli q bep3 swaps --page=2 --limit=100

```
kvcli query bep3 swaps [flags]
```

### Options

```
      --direction string    (optional) filter by atomic swap direction, direction: incoming/outgoing
      --expiration string   (optional) filter by atomic swaps that expire before a block height
      --height int          Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help                help for swaps
      --indent              Add indent to JSON response
      --involve string      (optional) filter by atomic swaps that involve an address
      --ledger              Use a connected Ledger device
      --limit int           pagination limit of atomic swaps to query for (default 100)
      --node string         <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --page int            pagination page of atomic swaps to to query for (default 1)
      --status string       (optional) filter by atomic swap status, status: open/completed/expired
      --trust-node          Trust connected full node (don't verify proofs for responses)
```

### Options inherited from parent commands

```
      --chain-id string   Chain ID of tendermint node
```

