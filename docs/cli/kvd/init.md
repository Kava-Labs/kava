<!--
title: init
-->
## kvd init

Initialize private validator, p2p, genesis, and application configuration files

### Synopsis

Initialize validators's and node's configuration files.

```
kvd init [moniker] [flags]
```

### Options

```
      --chain-id string   genesis file chain-id, if left blank will be randomly created
  -h, --help              help for init
      --home string       node's home directory (default "/Users/ruaridh/.kvd")
  -o, --overwrite         overwrite the genesis.json file
```

### Options inherited from parent commands

```
      --inv-check-period uint   Assert registered invariants every N blocks
      --log_level string        Log level (default "main:info,state:info,*:error")
```

