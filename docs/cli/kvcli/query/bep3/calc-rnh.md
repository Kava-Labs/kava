<!--
title: calc-rnh
-->
## kvcli query bep3 calc-rnh

calculates an example random number hash from an optional timestamp

```
kvcli query bep3 calc-rnh [unix-timestamp] [flags]
```

### Examples

```
bep3 calc-rnh now
```

### Options

```
      --height int    Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help          help for calc-rnh
      --indent        Add indent to JSON response
      --ledger        Use a connected Ledger device
      --node string   <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --trust-node    Trust connected full node (don't verify proofs for responses)
```

### Options inherited from parent commands

```
      --chain-id string   Chain ID of tendermint node
```

