<!--
title: proposals
-->
## kvcli query gov proposals

Query proposals with optional filters

### Synopsis

Query for a all paginated proposals that match optional filters:

Example:
$ kvcli query gov proposals --depositor cosmos1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk
$ kvcli query gov proposals --voter cosmos1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk
$ kvcli query gov proposals --status (DepositPeriod|VotingPeriod|Passed|Rejected)
$ kvcli query gov proposals --page=2 --limit=100

```
kvcli query gov proposals [flags]
```

### Options

```
      --depositor string   (optional) filter by proposals deposited on by depositor
      --height int         Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help               help for proposals
      --indent             Add indent to JSON response
      --ledger             Use a connected Ledger device
      --limit int          pagination limit of proposals to query for (default 100)
      --node string        <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
      --page int           pagination page of proposals to to query for (default 1)
      --status string      (optional) filter proposals by proposal status, status: deposit_period/voting_period/passed/rejected
      --trust-node         Trust connected full node (don't verify proofs for responses)
      --voter string       (optional) filter by proposals voted on by voted
```

### Options inherited from parent commands

```
      --chain-id string   Chain ID of tendermint node
```

