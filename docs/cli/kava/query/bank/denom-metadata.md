<!--
title: denom-metadata
-->
## kava query bank denom-metadata

Query the client metadata for coin denominations

### Synopsis

Query the client metadata for all the registered coin denominations

Example:
  To query for the client metadata of all coin denominations use:
  $ kava query bank denom-metadata

To query for the client metadata of a specific coin denomination use:
  $ kava query bank denom-metadata --denom=[denom]

```
kava query bank denom-metadata [flags]
```

### Options

```
      --denom string    The specific denomination to query client metadata for
      --height int      Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help            help for denom-metadata
      --node string     <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
  -o, --output string   Output format (text|json) (default "text")
```

### Options inherited from parent commands

```
      --chain-id string   The network chain ID
```

