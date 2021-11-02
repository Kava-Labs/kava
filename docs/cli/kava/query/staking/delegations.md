<!--
title: delegations
-->
## kava query staking delegations

Query all delegations made by one delegator

### Synopsis

Query delegations for an individual delegator on all validators.

Example:
$ kava query staking delegations kava1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p

```
kava query staking delegations [delegator-addr] [flags]
```

### Options

```
      --count-total       count total number of records in delegations to query for
      --height int        Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help              help for delegations
      --limit uint        pagination limit of delegations to query for (default 100)
      --node string       <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --offset uint       pagination offset of delegations to query for
  -o, --output string     Output format (text|json) (default "text")
      --page uint         pagination page of delegations to query for. This sets offset to a multiple of limit (default 1)
      --page-key string   pagination page-key of delegations to query for
      --reverse           results are sorted in descending order
```

### Options inherited from parent commands

```
      --chain-id string   The network chain ID
```

