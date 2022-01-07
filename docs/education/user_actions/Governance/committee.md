<!--
title: committee
-->
## kvcli tx gov submit-proposal committee

Submit a governance proposal to change a committee.

### Synopsis

Submit a governance proposal to create, alter, or delete a committee.

The proposal file must be the json encoded form of the proposal type you want to submit.
For example, to create or update a committee:
```
{
  "type": "kava/CommitteeChangeProposal",
  "value": {
    "title": "A Title",
    "description": "A description of this proposal.",
    "new_committee": {
      "type": "kava/MemberCommittee",
      "value": {
        "base_committee": {
          "id": "1",
          "description": "The description of this committee.",
          "members": [
            "kava182k5kyx9sy4ap624gm9gukr3jfj7fdx8jzrdgq"
          ],
          "permissions": [
            {
              "type": "kava/SimpleParamChangePermission",
              "value": {
                "allowed_params": [
                  {
                    "subspace": "cdp",
                    "key": "CircuitBreaker"
                  }
                ]
              }
            }
          ],
          "vote_threshold": "0.800000000000000000",
          "proposal_duration": "604800000000000",
          "tally_option": "FirstPastThePost"
        }
      }
    }
  }
}
```
and to delete a committee:
```
{
  "type": "kava/CommitteeDeleteProposal",
  "value": {
    "title": "A Title",
    "description": "A description of this proposal.",
    "committee_id": "1"
  }
}
```

```
kvcli tx gov submit-proposal committee [proposal-file] [deposit] [flags]
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
  -h, --help                     help for committee
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

