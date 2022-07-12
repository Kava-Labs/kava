# Concepts

## Vaults

Users only interact with vaults.

* Unlike yearn, vaults do not have any idle assets. The module does not hold
  assets other than temporarily when the user first transfers to the module
  account, then is supplied.

## Strategies

Vaults have a strategy to optimize yields with the corresponding vault asset.
