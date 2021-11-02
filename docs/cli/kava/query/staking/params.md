<!--
title: params
-->
## kava query staking params

Query the current staking parameters information

### Synopsis

Query values set as staking parameters.

Example:
$ kava query staking params

```
kava query staking params [flags]
```

### Options

```
      --height int      Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help            help for params
      --node string     <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
  -o, --output string   Output format (text|json) (default "text")
```

### Options inherited from parent commands

```
      --chain-id string   The network chain ID
```

