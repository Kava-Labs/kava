<!--
title: total
-->
## kvcli query supply total

Query the total supply of coins of the chain

### Synopsis

Query total supply of coins that are held by accounts in the
			chain.

Example:
$ kvcli query supply total

To query for the total supply of a specific coin denomination use:
$ kvcli query supply total stake

```
kvcli query supply total [denom] [flags]
```

### Options

```
      --height int    Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help          help for total
      --indent        Add indent to JSON response
      --ledger        Use a connected Ledger device
      --node string   <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --trust-node    Trust connected full node (don't verify proofs for responses)
```

### Options inherited from parent commands

```
      --chain-id string   Chain ID of tendermint node
```

