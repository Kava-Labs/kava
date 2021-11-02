<!--
title: rewards
-->
## kava query distribution rewards

Query all distribution delegator rewards or rewards from a particular validator

### Synopsis

Query all rewards earned by a delegator, optionally restrict to rewards from a single validator.

Example:
$ kava query distribution rewards kava1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p
$ kava query distribution rewards kava1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p kavavaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj

```
kava query distribution rewards [delegator-addr] [validator-addr] [flags]
```

### Options

```
      --height int      Use a specific height to query state at (this can error if the node is pruning state)
  -h, --help            help for rewards
      --node string     <host>:<port> to Tendermint RPC interface for this chain (default "tcp://localhost:26657")
  -o, --output string   Output format (text|json) (default "text")
```

### Options inherited from parent commands

```
      --chain-id string   The network chain ID
```

