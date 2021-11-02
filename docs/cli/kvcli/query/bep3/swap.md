<!--
title: swap
-->
## kvcli query bep3 swap

get atomic swap information

```
kvcli query bep3 swap [swap-id] [flags]
```

### Examples

```
bep3 swap 6682c03cc3856879c8fb98c9733c6b0c30758299138166b6523fe94628b1d3af
```

### Options

```
      --height int    Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help          help for swap
      --indent        Add indent to JSON response
      --ledger        Use a connected Ledger device
      --node string   <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --trust-node    Trust connected full node (don't verify proofs for responses)
```

### Options inherited from parent commands

```
      --chain-id string   Chain ID of tendermint node
```

