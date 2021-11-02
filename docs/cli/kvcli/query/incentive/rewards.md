<!--
title: rewards
-->
## kvcli query incentive rewards

query claimable rewards

### Synopsis

Query rewards with optional flags for owner and type

			Example:
			$ kvcli query incentive rewards
			$ kvcli query incentive rewards --owner kava15qdefkmwswysgg4qxgqpqr35k3m49pkx2jdfnw
			$ kvcli query incentive rewards --type hard
			$ kvcli query incentive rewards --type usdx-minting
			$ kvcli query incentive rewards --type delegator
			$ kvcli query incentive rewards --type swap
			$ kvcli query incentive rewards --type hard --owner kava15qdefkmwswysgg4qxgqpqr35k3m49pkx2jdfnw
			$ kvcli query incentive rewards --type hard --unsynced

```
kvcli query incentive rewards [flags]
```

### Options

```
      --height int     Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help           help for rewards
      --indent         Add indent to JSON response
      --ledger         Use a connected Ledger device
      --limit int      pagination limit of rewards to query for (default 100)
      --node string    <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --owner string   (optional) filter by owner address
      --page int       pagination page rewards of to to query for (default 1)
      --trust-node     Trust connected full node (don't verify proofs for responses)
      --type string    (optional) filter by a reward type: delegator|hard|usdx-minting|swap
      --unsynced       (optional) get unsynced claims
```

### Options inherited from parent commands

```
      --chain-id string   Chain ID of tendermint node
```

