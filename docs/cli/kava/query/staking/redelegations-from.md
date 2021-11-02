<!--
title: redelegations-from
-->
## kava query staking redelegations-from

Query all outgoing redelegatations from a validator

### Synopsis

Query delegations that are redelegating _from_ a validator.

Example:
$ kava query staking redelegations-from kavavaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj

```
kava query staking redelegations-from [validator-addr] [flags]
```

### Options

```
      --count-total       count total number of records in validator redelegations to query for
      --height int        Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help              help for redelegations-from
      --limit uint        pagination limit of validator redelegations to query for (default 100)
      --node string       <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --offset uint       pagination offset of validator redelegations to query for
  -o, --output string     Output format (text|json) (default "text")
      --page uint         pagination page of validator redelegations to query for. This sets offset to a multiple of limit (default 1)
      --page-key string   pagination page-key of validator redelegations to query for
      --reverse           results are sorted in descending order
```

### Options inherited from parent commands

```
      --chain-id string   The network chain ID
```

