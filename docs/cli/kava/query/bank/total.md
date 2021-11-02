<!--
title: total
-->
## kava query bank total

Query the total supply of coins of the chain

### Synopsis

Query total supply of coins that are held by accounts in the chain.

Example:
  $ kava query bank total

To query for the total supply of a specific coin denomination use:
  $ kava query bank total --denom=[denom]

```
kava query bank total [flags]
```

### Options

```
      --denom string    The specific balance denomination to query for
      --height int      Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help            help for total
      --node string     <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
  -o, --output string   Output format (text|json) (default "text")
```

### Options inherited from parent commands

```
      --chain-id string   The network chain ID
```

