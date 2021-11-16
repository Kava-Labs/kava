## A Quick Overview 
short on time? Read these and get a refresher of the most commont commands.

### Get a list of all keys
```
kvcli keys list 
```

### Query Account Tokens
```
kvcli q accounts <kava-addr>
```

### View Staking Delegations 
```
kvcli q staking delegations <kava-addr>
```

### Claim staking rewards 
```
kvcli tx distribution withdraw-all-rewards --gas 650000 --gas-prices 0.01ukava --from <kava-addr or key name>
```

### Querying outstanding staking rewards
```
kvcli q distribution rewards <kava-addr>
```

### Claim staking from individual validator
```
kvcli tx distribution withdraw-rewards <kava-validator-addre (starts with kavavaloper)> --gas 650000 --gas-prices 0.01ukava --from <kava-addr or key name>
```

### Claiming KAVA Delegator rewards (yielding HARD and SWP)
```
kvcli tx incentive claim-delegator --multiplier hard=large,swp=large --from <kava-addr or key name> --gas 800000 --gas-prices 0.01ukava 
```

### Querying outstanding HARD rewards
```
kvcli q incentive rewards --type delegator --owner <kava-address>
```

### View HARD Incentives 
```
kvcli q incentive rewards --type hard --owner <kava-address>
```

### Claim HARD Incentives 
```
kvcli tx incentive claim-hard --multiplier hard=large --multiplier ukava=large --from <kava-addr or key name> --gas 800000 --gas-prices 0.01ukava
```

### Send Coins 
```
kvcli tx send <key_name> <kava-account> <amount (100ukava for example)> --memo [memo] --gas-prices 0.01ukava
```

### Query Transaction Status Using TX Hash 
```
kvcli q tx <tx_hash>
```

### Adding a Key to Ledger 
```
kvcli keys add <key-name> --account <account-index> --legacy-hd-path --ledger
```