<!--
title: raw-params
-->
## kvcli query committee raw-params

Query raw parameter values from any module.

### Synopsis

Query the byte value of any module's parameters. Useful in debugging and verifying governance proposals.

```
kvcli query committee raw-params [subspace] [key] [flags]
```

### Examples

```
kvcli query committee raw-params cdp CollateralParams
```

### Options

```
      --height int    Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help          help for raw-params
      --indent        Add indent to JSON response
      --ledger        Use a connected Ledger device
      --node string   <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --trust-node    Trust connected full node (don't verify proofs for responses)
```

### Options inherited from parent commands

```
      --chain-id string   Chain ID of tendermint node
```

