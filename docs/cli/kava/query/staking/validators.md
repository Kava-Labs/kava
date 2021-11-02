<!--
title: validators
-->
## kava query staking validators

Query for all validators

### Synopsis

Query details about all validators on a network.

Example:
$ kava query staking validators

```
kava query staking validators [flags]
```

### Options

```
      --count-total       count total number of records in validators to query for
      --height int        Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help              help for validators
      --limit uint        pagination limit of validators to query for (default 100)
      --node string       <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --offset uint       pagination offset of validators to query for
  -o, --output string     Output format (text|json) (default "text")
      --page uint         pagination page of validators to query for. This sets offset to a multiple of limit (default 1)
      --page-key string   pagination page-key of validators to query for
      --reverse           results are sorted in descending order
```

### Options inherited from parent commands

```
      --chain-id string   The network chain ID
```

