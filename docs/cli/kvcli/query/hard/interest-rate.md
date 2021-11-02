<!--
title: interest-rate
-->
## kvcli query hard interest-rate

get current money market interest rates

### Synopsis

get current money market interest rates:

		Example:
		$ kvcli q hard interest-rate
		$ kvcli q hard interest-rate --denom bnb

```
kvcli query hard interest-rate [flags]
```

### Options

```
      --denom string   (optional) filter interest rates by denom
      --height int     Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help           help for interest-rate
      --indent         Add indent to JSON response
      --ledger         Use a connected Ledger device
      --node string    <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --trust-node     Trust connected full node (don't verify proofs for responses)
```

### Options inherited from parent commands

```
      --chain-id string   Chain ID of tendermint node
```

