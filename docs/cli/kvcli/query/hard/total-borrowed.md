<!--
title: total-borrowed
-->
## kvcli query hard total-borrowed

get total current borrowed amount

### Synopsis

get the total amount of coins currently borrowed using flags:

		Example:
		$ kvcli q hard total-borrowed
		$ kvcli q hard total-borrowed --denom bnb

```
kvcli query hard total-borrowed [flags]
```

### Options

```
      --denom string   (optional) filter total borrowed coins by denom
      --height int     Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help           help for total-borrowed
      --indent         Add indent to JSON response
      --ledger         Use a connected Ledger device
      --node string    <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --trust-node     Trust connected full node (don't verify proofs for responses)
```

### Options inherited from parent commands

```
      --chain-id string   Chain ID of tendermint node
```

