<!--
title: assert-invariants
-->
## kvd assert-invariants

Validates that the input genesis file is valid and invariants pass

### Synopsis

Reads the input genesis file into a genesis document, checks that the state is valid and asserts that all invariants pass.

```
kvd assert-invariants [genesis-file] [flags]
```

### Examples

```
kvd assert-invariants /path/to/genesis.json
```

### Options

```
  -h, --help   help for assert-invariants
```

### Options inherited from parent commands

```
      --inv-check-period uint   Assert registered invariants every N blocks
      --log_level string        Log level (default "main:info,state:info,*:error")
```

