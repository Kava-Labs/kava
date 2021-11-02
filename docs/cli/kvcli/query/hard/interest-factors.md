<!--
title: interest-factors
-->
## kvcli query hard interest-factors

get current global interest factors

### Synopsis

get current global interest factors:

		Example:
		$ kvcli q hard interest-factors
		$ kvcli q hard interest-factors --denom bnb

```
kvcli query hard interest-factors [flags]
```

### Options

```
      --denom string   (optional) filter interest factors by denom
      --height int     Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help           help for interest-factors
      --indent         Add indent to JSON response
      --ledger         Use a connected Ledger device
      --node string    <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --trust-node     Trust connected full node (don't verify proofs for responses)
```

### Options inherited from parent commands

```
      --chain-id string   Chain ID of tendermint node
```

