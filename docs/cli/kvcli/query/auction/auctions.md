<!--
title: auctions
-->
## kvcli query auction auctions

query auctions with optional filters

### Synopsis

Query for all paginated auctions that match optional filters:
Example:
$ kvcli q auction auctions --type=(collateral|surplus|debt)
$ kvcli q auction auctions --owner=kava1hatdq32u5x4wnxrtv5wzjzmq49sxgjgsj0mffm
$ kvcli q auction auctions --denom=bnb
$ kvcli q auction auctions --phase=(forward|reverse)
$ kvcli q auction auctions --page=2 --limit=100

```
kvcli query auction auctions [flags]
```

### Options

```
      --denom string   (optional) filter by auction denom
      --height int     Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help           help for auctions
      --indent         Add indent to JSON response
      --ledger         Use a connected Ledger device
      --limit int      pagination limit of auctions to query for (default 100)
      --node string    <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --owner string   (optional) filter by collateral auction owner
      --page int       pagination page of auctions to to query for (default 1)
      --phase string   (optional) filter by collateral auction phase, phase: forward/reverse
      --trust-node     Trust connected full node (don't verify proofs for responses)
      --type string    (optional) filter by auction type, type: collateral, debt, surplus
```

### Options inherited from parent commands

```
      --chain-id string   Chain ID of tendermint node
```

