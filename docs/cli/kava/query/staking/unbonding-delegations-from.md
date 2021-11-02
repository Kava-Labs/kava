<!--
title: unbonding-delegations-from
-->
## kava query staking unbonding-delegations-from

Query all unbonding delegatations from a validator

### Synopsis

Query delegations that are unbonding _from_ a validator.

Example:
$ kava query staking unbonding-delegations-from kavavaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj

```
kava query staking unbonding-delegations-from [validator-addr] [flags]
```

### Options

```
      --count-total       count total number of records in unbonding delegations to query for
      --height int        Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help              help for unbonding-delegations-from
      --limit uint        pagination limit of unbonding delegations to query for (default 100)
      --node string       <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --offset uint       pagination offset of unbonding delegations to query for
  -o, --output string     Output format (text|json) (default "text")
      --page uint         pagination page of unbonding delegations to query for. This sets offset to a multiple of limit (default 1)
      --page-key string   pagination page-key of unbonding delegations to query for
      --reverse           results are sorted in descending order
```

### Options inherited from parent commands

```
      --chain-id string   The network chain ID
```

