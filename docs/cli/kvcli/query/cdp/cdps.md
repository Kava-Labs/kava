<!--
title: cdps
-->
## kvcli query cdp cdps

query cdps with optional filters

### Synopsis

Query for all paginated cdps that match optional filters:
Example:
$ kvcli q cdp cdps --collateral-type=bnb
$ kvcli q cdp cdps --owner=kava1hatdq32u5x4wnxrtv5wzjzmq49sxgjgsj0mffm
$ kvcli q cdp cdps --id=21
$ kvcli q cdp cdps --ratio=2.75
$ kvcli q cdp cdps --page=2 --limit=100

```
kvcli query cdp cdps [flags]
```

### Options

```
      --collateral-type string   (optional) filter by CDP collateral type
      --height int               Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help                     help for cdps
      --id string                (optional) filter by CDP ID
      --indent                   Add indent to JSON response
      --ledger                   Use a connected Ledger device
      --limit int                pagination limit of CDPs to query for (default 100)
      --node string              <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --owner string             (optional) filter by CDP owner
      --page int                 pagination page of CDPs to to query for (default 1)
      --ratio string             (optional) filter by CDP collateralization ratio threshold
      --trust-node               Trust connected full node (don't verify proofs for responses)
```

### Options inherited from parent commands

```
      --chain-id string   Chain ID of tendermint node
```

