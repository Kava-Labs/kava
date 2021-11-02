<!--
title: tally
-->
## kvcli query committee tally

Get the current tally of votes on a proposal

### Synopsis

Query the current tally of votes on a proposal to see the progress of the voting.

```
kvcli query committee tally [proposal-id] [flags]
```

### Examples

```
kvcli query committee tally 2
```

### Options

```
      --height int    Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help          help for tally
      --indent        Add indent to JSON response
      --ledger        Use a connected Ledger device
      --node string   <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --trust-node    Trust connected full node (don't verify proofs for responses)
```

### Options inherited from parent commands

```
      --chain-id string   Chain ID of tendermint node
```

