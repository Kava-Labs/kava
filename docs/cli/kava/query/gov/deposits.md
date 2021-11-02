<!--
title: deposits
-->
## kava query gov deposits

Query deposits on a proposal

### Synopsis

Query details for all deposits on a proposal.
You can find the proposal-id by running "kava query gov proposals".

Example:
$ kava query gov deposits 1

```
kava query gov deposits [proposal-id] [flags]
```

### Options

```
      --count-total       count total number of records in deposits to query for
      --height int        Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help              help for deposits
      --limit uint        pagination limit of deposits to query for (default 100)
      --node string       <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --offset uint       pagination offset of deposits to query for
  -o, --output string     Output format (text|json) (default "text")
      --page uint         pagination page of deposits to query for. This sets offset to a multiple of limit (default 1)
      --page-key string   pagination page-key of deposits to query for
      --reverse           results are sorted in descending order
```

### Options inherited from parent commands

```
      --chain-id string   The network chain ID
```

