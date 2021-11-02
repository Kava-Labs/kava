<!--
title: historical-info
-->
## kvcli query staking historical-info

Query historical info at given height

### Synopsis

Query historical info at given height.

Example:
$ kvcli query staking historical-info 5

```
kvcli query staking historical-info [height] [flags]
```

### Options

```
      --height int    Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help          help for historical-info
      --indent        Add indent to JSON response
      --ledger        Use a connected Ledger device
      --node string   <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --trust-node    Trust connected full node (don't verify proofs for responses)
```

### Options inherited from parent commands

```
      --chain-id string   Chain ID of tendermint node
```

