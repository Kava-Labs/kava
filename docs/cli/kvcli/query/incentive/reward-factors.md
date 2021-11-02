<!--
title: reward-factors
-->
## kvcli query incentive reward-factors

get current global reward factors

### Synopsis

Get current global reward factors for all reward types.

```
kvcli query incentive reward-factors [flags]
```

### Options

```
      --denom string   (optional) filter reward factors by denom
      --height int     Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help           help for reward-factors
      --indent         Add indent to JSON response
      --ledger         Use a connected Ledger device
      --node string    <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --trust-node     Trust connected full node (don't verify proofs for responses)
```

### Options inherited from parent commands

```
      --chain-id string   Chain ID of tendermint node
```

