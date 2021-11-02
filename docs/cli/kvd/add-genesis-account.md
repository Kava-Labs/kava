<!--
title: add-genesis-account
-->
## kvd add-genesis-account

Add a genesis account to genesis.json

### Synopsis

Add a genesis account to genesis.json. The provided account must specify
 the account address or key name and a list of initial coins. If a key name is given,
 the address will be looked up in the local Keybase. The list of initial tokens must
 contain valid denominations. Accounts may optionally be supplied with vesting parameters.
 If the account is a periodic or validator vesting account, vesting periods must be supplied
 via a JSON file using the 'vesting-periods-file' flag or 'validator-vesting-file' flag,
 respectively.
 Example:
 kvcli add-genesis-account <account-name> <amount> --vesting-amount <amount> --vesting-end-time <unix-timestamp> --vesting-start-time <unix-timestamp> --vesting-periods <path/to/vesting.json>

```
kvd add-genesis-account [address_or_key_name] [coin][,[coin]] [flags]
```

### Options

```
  -h, --help                            help for add-genesis-account
      --home string                     node's home directory (default "/Users/ruaridh/.kvd")
      --home-client string              client's home directory (default "/Users/ruaridh/.kvcli")
      --keyring-backend string          Select keyring's backend (os|file|test) (default "os")
      --validator-vesting-file string   path to file where validator vesting schedule is specified
      --vesting-amount string           amount of coins for vesting accounts
      --vesting-end-time uint           schedule end time (unix epoch) for vesting accounts
      --vesting-periods-file string     path to file where periodic vesting schedule is specified
      --vesting-start-time uint         schedule start time (unix epoch) for vesting accounts
```

### Options inherited from parent commands

```
      --inv-check-period uint   Assert registered invariants every N blocks
      --log_level string        Log level (default "main:info,state:info,*:error")
```

