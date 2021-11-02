<!--
title: vote
-->
## kava query gov vote

Query details of a single vote

### Synopsis

Query details for a single vote on a proposal given its identifier.

Example:
$ kava query gov vote 1 cosmos1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk

```
kava query gov vote [proposal-id] [voter-addr] [flags]
```

### Options

```
      --height int      Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help            help for vote
      --node string     <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
  -o, --output string   Output format (text|json) (default "text")
```

### Options inherited from parent commands

```
      --chain-id string   The network chain ID
```

