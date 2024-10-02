Genesis Updates
===============
- The genesis.json account in this directory mimics the values that are present in `kava-internal-testnet` with the exception of `chain-id` and `validator` adjustments
- This file represents the initial state of the node when the validator node starts up in kava with requisite accounts, balances, permissions, etc
- These changes are needed for the `seed-protonet.sh` script to successfully execute in GitHub actions when changes are made to this branch, to reset initial node state
