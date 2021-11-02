<!--
title: migrate
-->
## kvd migrate

Migrate genesis file from kava v0.14 to v0.15

### Synopsis

Migrate the source genesis into the current version, sorts it, and print to STDOUT.

```
kvd migrate [genesis-file] [flags]
```

### Examples

```
kvd migrate /path/to/genesis.json
```

### Options

```
  -h, --help   help for migrate
```

### Options inherited from parent commands

```
      --inv-check-period uint   Assert registered invariants every N blocks
      --log_level string        Log level (default "main:info,state:info,*:error")
```

