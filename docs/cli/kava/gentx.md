<!--
title: gentx
-->
## kava gentx

Generate a genesis tx carrying a self delegation

### Synopsis

Generate a genesis transaction that creates a validator with a self-delegation,
that is signed by the key in the Keyring referenced by a given name. A node ID and Bech32 consensus
pubkey may optionally be provided. If they are omitted, they will be retrieved from the priv_validator.json
file. The following default parameters are included:
    
	delegation amount:           100000000stake
	commission rate:             0.1
	commission max rate:         0.2
	commission max change rate:  0.01
	minimum self delegation:     1


Example:
$ kava gentx my-key-name 1000000stake --home=/path/to/home/dir --keyring-backend=os --chain-id=test-chain-1 \
    --moniker="myvalidator" \
    --commission-max-change-rate=0.01 \
    --commission-max-rate=1.0 \
    --commission-rate=0.07 \
    --details="..." \
    --security-contact="..." \
    --website="..."


```
kava gentx [key_name] [amount] [flags]
```

### Options

```
  -a, --account-number uint                 The account number of the signing account (offline mode only)
      --amount string                       Amount of coins to bond
  -b, --broadcast-mode string               Transaction broadcasting mode (sync|async|block) (default "sync")
      --chain-id string                     The network chain ID
      --commission-max-change-rate string   The maximum commission change rate percentage (per day)
      --commission-max-rate string          The maximum commission rate percentage
      --commission-rate string              The initial commission rate percentage
      --details string                      The validator's (optional) details
      --dry-run                             ignore the --gas flag and perform a simulation of a transaction, but don't broadcast it
      --fee-account string                  Fee account pays fees for the transaction instead of deducting from the signer
      --fees string                         Fees to pay along with transaction; eg: 10uatom
      --from string                         Name or address of private key with which to sign
      --gas string                          gas limit to set per-transaction; set to "auto" to calculate sufficient gas automatically (default 200000)
      --gas-adjustment float                adjustment factor to be multiplied against the estimate returned by the tx simulation; if the gas limit is set manually this flag is ignored  (default 1)
      --gas-prices string                   Gas prices in decimal format to determine the transaction fee (e.g. 0.1uatom)
      --generate-only                       Build an unsigned transaction and write it to STDOUT (when enabled, the local Keybase is not accessible)
  -h, --help                                help for gentx
      --home string                         The application home directory
      --identity string                     The (optional) identity signature (ex. UPort or Keybase)
      --ip string                           The node's public IP (default "192.168.1.248")
      --keyring-backend string              Select keyring's backend (os|file|kwallet|pass|test|memory) (default "os")
      --keyring-dir string                  The client Keyring directory; if omitted, the default 'home' directory will be used
      --ledger                              Use a connected Ledger device
      --min-self-delegation string          The minimum self delegation required on the validator
      --moniker string                      The validator's (optional) moniker
      --node string                         <host>:<port> to tendermint rpc interface for this chain (default "tcp://localhost:26657")
      --node-id string                      The node's NodeID
      --note string                         Note to add a description to the transaction (previously --memo)
      --offline                             Offline mode (does not allow any online functionality
  -o, --output string                       Output format (text|json) (default "json")
      --output-document string              Write the genesis transaction JSON document to the given file instead of the default location
      --pubkey string                       The validator's Protobuf JSON encoded public key
      --security-contact string             The validator's (optional) security contact email
  -s, --sequence uint                       The sequence number of the signing account (offline mode only)
      --sign-mode string                    Choose sign mode (direct|amino-json), this is an advanced feature
      --timeout-height uint                 Set a block timeout height to prevent the tx from being committed past a certain height
      --website string                      The validator's (optional) website
  -y, --yes                                 Skip tx broadcasting prompt confirmation
```

