<!--
title: votes
-->
## kvcli query gov votes

Query votes on a proposal

### Synopsis

Query vote details for a single proposal by its identifier.

Example:
$ kvcli query gov votes 1
$ kvcli query gov votes 1 --page=2 --limit=100

```
kvcli query gov votes [proposal-id] [flags]
```

### Options

```
      --height int    Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help          help for votes
      --indent        Add indent to JSON response
      --ledger        Use a connected Ledger device
      --limit int     pagination limit of votes to query for (default 100)
      --node string   <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --page int      pagination page of votes to to query for (default 1)
      --trust-node    Trust connected full node (don't verify proofs for responses)
```

### Options inherited from parent commands

```
      --chain-id string   Chain ID of tendermint node
```

