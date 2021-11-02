<!--
title: deposit
-->
## kava query gov deposit

Query details of a deposit

### Synopsis

Query details for a single proposal deposit on a proposal by its identifier.

Example:
$ kava query gov deposit 1 cosmos1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk

```
kava query gov deposit [proposal-id] [depositer-addr] [flags]
```

### Options

```
      --height int      Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help            help for deposit
      --node string     <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
  -o, --output string   Output format (text|json) (default "text")
```

### Options inherited from parent commands

```
      --chain-id string   The network chain ID
```

