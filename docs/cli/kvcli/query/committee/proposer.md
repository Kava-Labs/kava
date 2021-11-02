<!--
title: proposer
-->
## kvcli query committee proposer

Query the proposer of a governance proposal

### Synopsis

Query which address proposed a proposal with a given ID.

```
kvcli query committee proposer [proposal-id] [flags]
```

### Examples

```
kvcli query committee proposer 2
```

### Options

```
      --height int    Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help          help for proposer
      --indent        Add indent to JSON response
      --ledger        Use a connected Ledger device
      --node string   <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --trust-node    Trust connected full node (don't verify proofs for responses)
```

### Options inherited from parent commands

```
      --chain-id string   Chain ID of tendermint node
```

