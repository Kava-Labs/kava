Genesis Updates
===============
- Added the God Committee section to the `app_state.committee.committees` section 
- Added dev wallet `kava1vlpsrmdyuywvaqrv7rx6xga224sqfwz3fyfhwq` to `app_state.auth.accounts` for genesis account
- Added dev wallet balances to `app_state.bank.balances` to enable this account to fund various other accounts
- Updated `app_state.issuance.params` with different kava assets that will be needed in `seed-protonet.sh` script

Summary
=======
- We decided to move critical data sections into the genesis.json file to ensure the node that starts up has all relevant state it needs BEFORE we execute any scripts to manipulate state.
- The genesis.json contains the initial dev wallet, genesis account and associated balances to said account which is then used to pre-fund various ancillary kava balances (ex: community) to serve whatever testing purpose to protonet
