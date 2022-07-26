# Concepts

## Vaults

Vaults are the only user facing interface. Users only need to deposit and
withdraw from a vault.

Vaults do not have any idle assets. The module does not hold assets other than
temporarily in the middle of a transaction when the user first transfers to the
module account, then is supplied.

## Strategies

Vaults have a strategy to optimize yields with the corresponding vault asset.

Currently, the available strategies are:

1. **Hard** - Deposits assets such as USDX to the Hard module
2. **Savings** - Deposits assets to the Savings module
