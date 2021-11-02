<!--
title: param
-->
## kava query gov param

Query the parameters (voting|tallying|deposit) of the governance process

### Synopsis

Query the all the parameters for the governance process.

Example:
$ kava query gov param voting
$ kava query gov param tallying
$ kava query gov param deposit

```
kava query gov param [param-type] [flags]
```

### Options

```
      --height int      Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help            help for param
      --node string     <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
  -o, --output string   Output format (text|json) (default "text")
```

### Options inherited from parent commands

```
      --chain-id string   The network chain ID
```

