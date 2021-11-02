<!--
title: deposits
-->
## kvcli query swap deposits

get liquidity provider deposits

### Synopsis

get liquidity provider deposits:
		Example:
		$ kvcli q swap deposits --pool bnb:usdx
		$ kvcli q swap deposits --owner kava1l0xsq2z7gqd7yly0g40y5836g0appumark77ny
		$ kvcli q swap deposits --pool bnb:usdx --owner kava1l0xsq2z7gqd7yly0g40y5836g0appumark77ny
		$ kvcli q swap deposits --page=2 --limit=100

```
kvcli query swap deposits [flags]
```

### Options

```
      --height int     Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help           help for deposits
      --indent         Add indent to JSON response
      --ledger         Use a connected Ledger device
      --limit int      pagination limit of deposits to query for (default 100)
      --node string    <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --owner string   owner, also known as a liquidity provider
      --page int       pagination page of deposits to query for (default 1)
      --pool string    pool name
      --trust-node     Trust connected full node (don't verify proofs for responses)
```

### Options inherited from parent commands

```
      --chain-id string   Chain ID of tendermint node
```

