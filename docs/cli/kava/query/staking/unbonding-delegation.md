<!--
title: unbonding-delegation
-->
## kava query staking unbonding-delegation

Query an unbonding-delegation record based on delegator and validator address

### Synopsis

Query unbonding delegations for an individual delegator on an individual validator.

Example:
$ kava query staking unbonding-delegation kava1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p kavavaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj

```
kava query staking unbonding-delegation [delegator-addr] [validator-addr] [flags]
```

### Options

```
      --height int      Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help            help for unbonding-delegation
      --node string     <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
  -o, --output string   Output format (text|json) (default "text")
```

### Options inherited from parent commands

```
      --chain-id string   The network chain ID
```

