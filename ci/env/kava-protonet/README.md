Genesis Updates
===============
- Added the God Committee section to the `app_state.committee.committees` section 
- Added dev wallet `kava1vlpsrmdyuywvaqrv7rx6xga224sqfwz3fyfhwq` to `app_state.auth.accounts` for genesis account
- Added dev wallet balances to `app_state.bank.balances` to enable this account to fund various other accounts
- Updated `app_state.issuance.params` with different kava assets that will be needed in `seed-protonet.sh` script

Summary
=======
- We decided to move critical bootstrapping sections into the genesis.json file to ensure the node that kava starts up has all relevant initial state it needs BEFORE the seed-protonet.sh script executes.

Advantages
- Single source of truth for initial provisioning of the kava state when the validator node starts up

Disadvantages
- Maybe less dynamic in terms of having a shell script, post kava install, setup genesis accounts etc on the fly. Not sure if this was an expectation of this project.

