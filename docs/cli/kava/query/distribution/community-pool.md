<!--
title: community-pool
-->
## kava query distribution community-pool

Query the amount of coins in the community pool

### Synopsis

Query all coins in the community pool which is under Governance control.

Example:
$ kava query distribution community-pool

```
kava query distribution community-pool [flags]
```

### Options

```
      --height int      Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help            help for community-pool
      --node string     <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
  -o, --output string   Output format (text|json) (default "text")
```

### Options inherited from parent commands

```
      --chain-id string   The network chain ID
```

