<!--
title: deposits
-->
## kvcli query cdp deposits

get deposits for a cdp

### Synopsis

Get the deposits of a CDP.

Example:
$ kvcli query cdp deposits kava15qdefkmwswysgg4qxgqpqr35k3m49pkx2jdfnw atom-a

```
kvcli query cdp deposits [owner-addr] [collateral-type] [flags]
```

### Options

```
      --height int    Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help          help for deposits
      --indent        Add indent to JSON response
      --ledger        Use a connected Ledger device
      --node string   <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --trust-node    Trust connected full node (don't verify proofs for responses)
```

### Options inherited from parent commands

```
      --chain-id string   Chain ID of tendermint node
```

