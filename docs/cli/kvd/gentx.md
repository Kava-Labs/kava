<!--
title: gentx
-->
## kvd gentx

Generate a genesis tx carrying a self delegation

### Synopsis

This command is an alias of the 'tx create-validator' command'.

		It creates a genesis transaction to create a validator. 
		The following default parameters are included: 
		    
	delegation amount:           100000000stake
	commission rate:             0.1
	commission max rate:         0.2
	commission max change rate:  0.01
	minimum self delegation:     1


```
kvd gentx [flags]
```

### Options

```
      --amount string                       Amount of coins to bond
      --commission-max-change-rate string   The maximum commission change rate percentage (per day)
      --commission-max-rate string          The maximum commission rate percentage
      --commission-rate string              The initial commission rate percentage
      --details string                      The validator's (optional) details
  -h, --help                                help for gentx
      --home string                         node's home directory (default "/Users/ruaridh/.kvd")
      --home-client string                  client's home directory (default "/Users/ruaridh/.kvcli")
      --identity string                     The (optional) identity signature (ex. UPort or Keybase)
      --ip string                           The node's public IP (default "192.168.1.248")
      --keyring-backend string              Select keyring's backend (os|file|test) (default "os")
      --min-self-delegation string          The minimum self delegation required on the validator
      --name string                         name of private key with which to sign the gentx
      --node-id string                      The node's NodeID
      --output-document string              write the genesis transaction JSON document to the given file instead of the default location
      --pubkey string                       The Bech32 encoded PubKey of the validator
      --security-contact string             The validator's (optional) security contact email
      --website string                      The validator's (optional) website
```

### Options inherited from parent commands

```
      --inv-check-period uint   Assert registered invariants every N blocks
      --log_level string        Log level (default "main:info,state:info,*:error")
```

