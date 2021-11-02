<!--
title: proposal
-->
## kvcli query gov proposal

Query details of a single proposal

### Synopsis

Query details for a proposal. You can find the
proposal-id by running "kvcli query gov proposals".

Example:
$ kvcli query gov proposal 1

```
kvcli query gov proposal [proposal-id] [flags]
```

### Options

```
      --height int    Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help          help for proposal
      --indent        Add indent to JSON response
      --ledger        Use a connected Ledger device
      --node string   <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --trust-node    Trust connected full node (don't verify proofs for responses)
```

### Options inherited from parent commands

```
      --chain-id string   Chain ID of tendermint node
```

