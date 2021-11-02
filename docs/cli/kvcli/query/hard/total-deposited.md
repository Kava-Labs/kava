<!--
title: total-deposited
-->
## kvcli query hard total-deposited

get total current deposited amount

### Synopsis

get the total amount of coins currently deposited using flags:

		Example:
		$ kvcli q hard total-deposited
		$ kvcli q hard total-deposited --denom bnb

```
kvcli query hard total-deposited [flags]
```

### Options

```
      --denom string   (optional) filter total deposited coins by denom
      --height int     Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help           help for total-deposited
      --indent         Add indent to JSON response
      --ledger         Use a connected Ledger device
      --node string    <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --trust-node     Trust connected full node (don't verify proofs for responses)
```

### Options inherited from parent commands

```
      --chain-id string   Chain ID of tendermint node
```

