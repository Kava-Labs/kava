<!--
title: community-pool-multi-spend
-->
## kvcli tx gov submit-proposal community-pool-multi-spend

Submit a community pool multi-spend proposal

### Synopsis

Submit a community pool multi-spend proposal along with an initial deposit.
The proposal details must be supplied via a JSON file.

Example:
$ kvcli tx gov submit-proposal community-pool-multi-spend <path/to/proposal.json> --from=<key_or_address>

Where proposal.json contains:

{
  "title": "Community Pool Multi-Spend",
  "description": "Pay many users some KAVA!",
  "recipient_list": [
		{
			"address": "kava1mz2003lathm95n5vnlthmtfvrzrjkrr53j4464",
			"amount": [
				{
					"denom": "ukava",
					"amount": "1000000"
				}
			]
		},
		{
			"address": "kava1zqezafa0luyetvtj8j67g336vaqtuudnsjq7vm",
			"amount": [
				{
					"denom": "ukava",
					"amount": "1000000"
				}
			]
		}
	],
	"deposit": [
		{
			"denom": "ukava",
			"amount": "1000000000"
		}
	]
}

```
kvcli tx gov submit-proposal community-pool-multi-spend [proposal-file] [flags]
```

### Options

```
  -a, --account-number uint      The account number of the signing account (offline mode only)
  -b, --broadcast-mode string    Transaction broadcasting mode (sync|async|block) (default "sync")
      --dry-run                  ignore the --gas flag and perform a simulation of a transaction, but don't broadcast it
      --fees string              Fees to pay along with transaction; eg: 10uatom
      --from string              Name or address of private key with which to sign
      --gas string               gas limit to set per-transaction; set to "auto" to calculate required gas automatically (default 200000) (default "200000")
      --gas-adjustment float     adjustment factor to be multiplied against the estimate returned by the tx simulation; if the gas limit is set manually this flag is ignored  (default 1)
      --gas-prices string        Gas prices to determine the transaction fee (e.g. 10uatom)
      --generate-only            Build an unsigned transaction and write it to STDOUT (when enabled, the local Keybase is not accessible and the node operates offline)
  -h, --help                     help for community-pool-multi-spend
      --indent                   Add indent to JSON response
      --keyring-backend string   Select keyring's backend (os|file|test) (default "os")
      --ledger                   Use a connected Ledger device
      --memo string              Memo to send along with transaction
      --node string              <host>:<port> to tendermint rpc interface for this chain (default "tcp://localhost:26657")
  -s, --sequence uint            The sequence number of the signing account (offline mode only)
      --trust-node               Trust connected full node (don't verify proofs for responses) (default true)
  -y, --yes                      Skip tx broadcasting prompt confirmation
```

### Options inherited from parent commands

```
      --chain-id string   Chain ID of tendermint node
```

