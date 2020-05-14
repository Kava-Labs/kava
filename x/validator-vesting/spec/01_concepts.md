# Concepts

The validator-vesting module is responsible for managing Validator Vesting Accounts, a vesting account for which the release of coins is tied to the validation of the blockchain. Validator Vesting Accounts implement the cosmos-sdk vesting account spec, which can be found [here](https://github.com/cosmos/cosmos-sdk/tree/master/x/auth/spec).

The main concept the Validator Vesting Account introduces is that of _conditional vesting_, or vesting accounts in which it is possible for some or all of the vesting coins to fail to vest. For Validator Vesting Accounts, vesting is broken down into user-specified __vesting periods__. Each vesting period specifies an amount of coins that  could vest in that period, and how long the vesting period lasts.

For each vesting period, a __signing threshold__ is specified, which is the percentage of blocks that must be signed for the coins to successfully vest. After a period ends, coins that are successfully vested become freely spendable. Coins that do not successfully vest are burned, or sent to an optional return address.
