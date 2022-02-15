<!--
order: 1
-->
# Quick Reference


### Get a list of all keys
```
kava keys list 
```

### Query Account 
```
kava q account <kava-addr>
```

### View Staking Delegations 
```
kava q staking delegations <kava-addr>
```

### Claim staking rewards 
```
kava tx distribution withdraw-all-rewards --gas 650000 --gas-prices 0.01ukava --from <kava-addr or key name>
```

### Querying outstanding staking rewards
```
kava q distribution rewards <kava-addr>
```

### Claim staking from individual validator
```
kava tx distribution withdraw-rewards <kava-validator-addre (starts with kavavaloper)> --gas 650000 --gas-prices 0.01ukava --from <kava-addr or key name>
```

### Claiming KAVA Delegator rewards (yielding HARD and SWP)
```
kava tx incentive claim-delegator --multiplier hard=large,swp=large --from <kava-addr or key name> --gas 800000 --gas-prices 0.01ukava 
```

### Querying outstanding HARD rewards
```
kava q incentive rewards --type delegator --owner <kava-address>
```

### View HARD Incentives 
```
kava q incentive rewards --type hard --owner <kava-address>
```

### Claim HARD Incentives 
```
kava tx incentive claim-hard --multiplier hard=large --multiplier ukava=large --from <kava-addr or key name> --gas 800000 --gas-prices 0.01ukava
```

### Send Coins 
```
kava tx bank send <from> <to> <amount (100ukava for example)> --memo [memo] --gas-prices 0.01ukava
```

### Query Transaction Status Using TX Hash 
```
kava q tx <tx_hash>
```

### Query Transactions Using events 
```
kava q txs --events message.actions=swap_deposit
```


### Adding a Key to Ledger 
```
kava keys add <key-name> --account <account-index> --legacy-hd-path --ledger
```