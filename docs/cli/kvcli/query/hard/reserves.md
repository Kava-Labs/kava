<!--
title: reserves
-->
## kvcli query hard reserves

get total current Hard module reserves

### Synopsis

get the total amount of coins currently held as reserve by the Hard module:

		Example:
		$ kvcli q hard reserves
		$ kvcli q hard reserves --denom bnb

```
kvcli query hard reserves [flags]
```

### Options

```
      --denom string   (optional) filter reserve coins by denom
      --height int     Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help           help for reserves
      --indent         Add indent to JSON response
      --ledger         Use a connected Ledger device
      --node string    <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --trust-node     Trust connected full node (don't verify proofs for responses)
```

### Options inherited from parent commands

```
      --chain-id string   Chain ID of tendermint node
```

