<!--
title: deposits
-->
## kvcli query hard deposits

query hard module deposits with optional filters

### Synopsis

query for all hard module deposits or a specific deposit using flags:

		Example:
		$ kvcli q hard deposits
		$ kvcli q hard deposits --owner kava1l0xsq2z7gqd7yly0g40y5836g0appumark77ny --denom bnb
		$ kvcli q hard deposits --denom ukava
		$ kvcli q hard deposits --denom btcb

```
kvcli query hard deposits [flags]
```

### Options

```
      --denom string   (optional) filter for deposits by denom
      --height int     Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help           help for deposits
      --indent         Add indent to JSON response
      --ledger         Use a connected Ledger device
      --limit int      pagination limit (max 100) (default 100)
      --node string    <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --owner string   (optional) filter for deposits by owner address
      --page int       pagination page to query for (default 1)
      --trust-node     Trust connected full node (don't verify proofs for responses)
```

### Options inherited from parent commands

```
      --chain-id string   Chain ID of tendermint node
```

