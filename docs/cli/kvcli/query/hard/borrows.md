<!--
title: borrows
-->
## kvcli query hard borrows

query hard module borrows with optional filters

### Synopsis

query for all hard module borrows or a specific borrow using flags:

		Example:
		$ kvcli q hard borrows
		$ kvcli q hard borrows --owner kava1l0xsq2z7gqd7yly0g40y5836g0appumark77ny
		$ kvcli q hard borrows --denom bnb

```
kvcli query hard borrows [flags]
```

### Options

```
      --denom string   (optional) filter for borrows by denom
      --height int     Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help           help for borrows
      --indent         Add indent to JSON response
      --ledger         Use a connected Ledger device
      --limit int      pagination limit (max 100) (default 100)
      --node string    <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --owner string   (optional) filter for borrows by owner address
      --page int       pagination page to query for (default 1)
      --trust-node     Trust connected full node (don't verify proofs for responses)
```

### Options inherited from parent commands

```
      --chain-id string   Chain ID of tendermint node
```

