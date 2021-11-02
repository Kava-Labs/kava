<!--
title: export
-->
## kvd export

Export state to JSON

```
kvd export [flags]
```

### Options

```
      --for-zero-height          Export state to start at height zero (perform preproccessing)
      --height int               Export state from a particular height (-1 means latest height) (default -1)
  -h, --help                     help for export
      --jail-whitelist strings   List of validators to not jail state export
```

### Options inherited from parent commands

```
      --inv-check-period uint   Assert registered invariants every N blocks
      --log_level string        Log level (default "main:info,state:info,*:error")
```

