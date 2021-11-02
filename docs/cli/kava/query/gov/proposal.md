<!--
title: proposal
-->
## kava query gov proposal

Query details of a single proposal

### Synopsis

Query details for a proposal. You can find the
proposal-id by running "kava query gov proposals".

Example:
$ kava query gov proposal 1

```
kava query gov proposal [proposal-id] [flags]
```

### Options

```
      --height int      Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help            help for proposal
      --node string     <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
  -o, --output string   Output format (text|json) (default "text")
```

### Options inherited from parent commands

```
      --chain-id string   The network chain ID
```

