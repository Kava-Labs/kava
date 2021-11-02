<!--
title: delete
-->
## kava keys delete

Delete the given keys

### Synopsis

Delete keys from the Keybase backend.

Note that removing offline or ledger keys will remove
only the public key references stored locally, i.e.
private keys stored in a ledger device cannot be deleted with the CLI.


```
kava keys delete <name>... [flags]
```

### Options

```
  -f, --force   Remove the key unconditionally without asking for the passphrase. Deprecated.
  -h, --help    help for delete
  -y, --yes     Skip confirmation prompt when deleting offline or ledger key references
```

### Options inherited from parent commands

```
      --home string              The application home directory
      --keyring-backend string   Select keyring's backend (os|file|test) (default "os")
      --keyring-dir string       The client Keyring directory; if omitted, the default 'home' directory will be used
      --output string            Output format (text|json) (default "text")
```

