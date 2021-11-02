<!--
title: cdp
-->
## kvcli query cdp cdp

get info about a cdp

### Synopsis

Get a CDP by the owner address and the collateral name.

Example:
$ kvcli query cdp cdp kava15qdefkmwswysgg4qxgqpqr35k3m49pkx2jdfnw atom-a

```
kvcli query cdp cdp [owner-addr] [collateral-type] [flags]
```

### Options

```
      --height int    Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help          help for cdp
      --indent        Add indent to JSON response
      --ledger        Use a connected Ledger device
      --node string   <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --trust-node    Trust connected full node (don't verify proofs for responses)
```

### Options inherited from parent commands

```
      --chain-id string   Chain ID of tendermint node
```

