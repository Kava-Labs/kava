<!--
title: historical-info
-->
## kava query staking historical-info

Query historical info at given height

### Synopsis

Query historical info at given height.

Example:
$ kava query staking historical-info 5

```
kava query staking historical-info [height] [flags]
```

### Options

```
      --height int      Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help            help for historical-info
      --node string     <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
  -o, --output string   Output format (text|json) (default "text")
```

### Options inherited from parent commands

```
      --chain-id string   The network chain ID
```

