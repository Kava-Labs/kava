<!--
title: evidence
order: 0
-->
## kvcli query evidence

Query for evidence by hash or for all (paginated) submitted evidence

### Synopsis

Query for specific submitted evidence by hash or query for all (paginated) evidence:
	
Example:
$ kvcli query evidence DF0C23E8634E480F84B9D5674A7CDC9816466DEC28A3358F73260F68D28D7660
$ kvcli query evidence --page=2 --limit=50

```
kvcli query evidence [flags]
```

### Options

```
      --height int    Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help          help for evidence
      --indent        Add indent to JSON response
      --ledger        Use a connected Ledger device
      --limit int     pagination limit of evidence to query for (default 100)
      --node string   <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --page int      pagination page of evidence to to query for (default 1)
      --trust-node    Trust connected full node (don't verify proofs for responses)
```

### Options inherited from parent commands

```
      --chain-id string   Chain ID of tendermint node
```

