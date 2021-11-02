<!--
title: delegation
-->
## kava query staking delegation

Query a delegation based on address and validator address

### Synopsis

Query delegations for an individual delegator on an individual validator.

Example:
$ kava query staking delegation kava1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p kavavaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj

```
kava query staking delegation [delegator-addr] [validator-addr] [flags]
```

### Options

```
      --height int      Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help            help for delegation
      --node string     <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
  -o, --output string   Output format (text|json) (default "text")
```

### Options inherited from parent commands

```
      --chain-id string   The network chain ID
```

