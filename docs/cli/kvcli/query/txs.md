<!--
title: txs
-->
## kvcli query txs

Query for paginated transactions that match a set of events

### Synopsis

Search for transactions that match the exact given events where results are paginated.
Each event takes the form of '{eventType}.{eventAttribute}={value}'. Please refer
to each module's documentation for the full set of events to query for. Each module
documents its respective events under 'xx_events.md'.

Example:
$ kvcli query txs --events 'message.sender=cosmos1...&message.action=withdraw_delegator_reward' --page 1 --limit 30

```
kvcli query txs [flags]
```

### Options

```
      --events string   list of transaction events in the form of {eventType}.{eventAttribute}={value}
  -h, --help            help for txs
      --limit uint32    Query number of transactions results per page returned (default 30)
  -n, --node string     Node to connect to (default "tcp://localhost:26657")
      --page uint32     Query a specific page of paginated results (default 1)
      --trust-node      Trust connected full node (don't verify proofs for responses)
```

### Options inherited from parent commands

```
      --chain-id string   Chain ID of tendermint node
```

