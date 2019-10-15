# Changelog

## [Unreleased]

### Features
[\#253] Add a new validator vesting account type, which releases coins on a periodic vesting schedule based on if a specific validator signed sufficient pre-commits. If the validtor didn't sign enough pre-commits, the vesting coins are burned or sent to a return address.
[\#260] Pin to cosmos-sdk commit #18de630 (tendermint 0.32.6)


### Improvements
[\#257] Include scripts to run large-scale simulations remotely using aws-batch
