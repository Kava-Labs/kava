## Transaction Creation

One of the fundamental ways to interact with a Kava Node is through the use of ```kvcli```. It can be used to create new transactions from the user, which will later be broadcasted to other nodes in the network after going through validations and signing. the ```kvcli``` commands for transaction creation will all follow a very similar pattern across modules and it will follow the structure below: 

```
kvcli tx [module_name] [command_name] [arguments] [flags]
```

| key    | explanation  |
|-----------|---|
| kvcli | calls the cli by its name |
| tx | creates a transaction |
| module_name | specifies the module which a transaction will be created |
| command_name | specifies the type of transaction for the given module |
| arguments | provide arguments for the transaction |
| flags | provide flags for the transaction some of which can be required |

the flags are the last part of the command and depending on the module they can be required in order for the transaction command to be accepted. Flags are reserved to modify commands, all flags will have a default value. Some of these default values will be empty strings and as stated above they will need to be filled in in order for the command to be processed. a user can override any flag's default value by ```--flag_name [flag_value]```. 

Users creating transactions will usually interact with the most common flags below: 


| Identifier       | Purpose |
|------------------|---------|
| --from           | transaction origin account |
| --gas            | how much gas a transaction consumes usually left as auto for estimation |
| --gas-adjustment | optional flag for scaling up the gas flag to avoid underestimating  |
| --gas-prices     | will set how much the user is willing to pay per gas unit |
| --fees           | will set the total fees the user is willing to pay |

the fees/gas prices that are calculated will decide if any given validator will decide to include the transaction in a new block by checking the value against their Node's minimum gas price. For that reason users are encouraged to pay more fees such that their transactions get added to a new block. 

### Transaction example 

given two accounts ```A``` & ```B```  the owner of account ```A``` would like to send some funds over to another account ```B``` , using the 	```kvcli``` the user with account ```A``` would create a transaction that will follow this structure below: 
```
kvcli tx send <B_Addr> 1000ukava --from <A_Addr> --gas-prices 0.50ukava
```

## Transaction Journey 

In the Cosmos SDK a transaction is defined as any user created object that must trigger state changes in the application. As such a transaction must be processed carefully and go through strict validation to ensure that it: 

1. Provides all the required fields.
2. The transaction is Valid. 
3. The user is authorized and has the fees to pay for the transaction. 
4. The transaction is not repeated or malicious in any way. 

### Mempool 
the first step for a transaction is for it to end up in the Mempool such that it can be processed by Validators and added to a new Block. In order for the transaction to be added into the Mempool it must pass a couple validation steps. 

1. Pass basic stateless checks that are quick in nature and act as a quick way to disqualify a transaction immediately, if it is failing to provide basic fields such as address, etc. 
2. Pass stateful checks which will required more thorough testing such as the address are valid, have enough funds, and are authorized to create the transaction. The reason this is stateful is because the node will create  the state copy and run the transaction in it to validate it. 

### Proposal 

Once a transaction is in the Mempool and a Proposer is chosen from the set of validators, the Proposer includes it in the new block. And a series of checks are ran by other validators to run the transactions and if everything is valid a new block will be created. 