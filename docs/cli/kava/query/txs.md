<!--
title: txs
-->
## kava query txs

Query for paginated transactions that match a set of events

### Synopsis

Search for transactions that match the exact given events where results are paginated.
Each event takes the form of '{eventType}.{eventAttribute}={value}'. Please refer
to each module's documentation for the full set of events to query for. Each module
documents its respective events under 'xx_events.md'.

Example:
$ kava query txs --events 'message.sender=cosmos1...&message.action=withdraw_delegator_reward' --page 1 --limit 30

```
kava query txs [flags]
```

### Options

```
      --events string   list of transaction events in the form of {eventType}.{eventAttribute}={value}
      --height int      Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help            help for txs
      --limit int       Query number of transactions results per page returned (default 30)
      --node string     <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
  -o, --output string   Output format (text|json) (default "text")
      --page int        Query a specific page of paginated results (default 1)
```

### Options inherited from parent commands

```
      --chain-id string   The network chain ID
```

