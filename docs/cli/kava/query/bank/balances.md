<!--
title: balances
-->
## kava query bank balances

Query for account balances by address

### Synopsis

Query the total balance of an account or of a specific denomination.

Example:
  $ kava query bank balances [address]
  $ kava query bank balances [address] --denom=[denom]

```
kava query bank balances [address] [flags]
```

### Options

```
      --count-total       count total number of records in all balances to query for
      --denom string      The specific balance denomination to query for
      --height int        Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help              help for balances
      --limit uint        pagination limit of all balances to query for (default 100)
      --node string       <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --offset uint       pagination offset of all balances to query for
  -o, --output string     Output format (text|json) (default "text")
      --page uint         pagination page of all balances to query for. This sets offset to a multiple of limit (default 1)
      --page-key string   pagination page-key of all balances to query for
      --reverse           results are sorted in descending order
```

### Options inherited from parent commands

```
      --chain-id string   The network chain ID
```

