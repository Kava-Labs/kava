# `validator-vesting`

<!-- TOC -->
1. **[Concepts](01_concepts.md)**
2. **[State](02_state.md)**
3. **[BeginBlock](03_begin_block.md)**

## Abstract

`x/validator-vesting` is an implementation of a Cosmos SDK sub-module that defines a new type of vesting account, `ValidatorVestingAccount`. This account implements the Cosmos SDK `VestingAccount` interface and extends it to add conditions to the vesting balance. In this implementation, in order to receive the vesting balance, the validator vesting account specifies a validator that must sign a given `SigningThreshold` of blocks during each vesting period in order for coins to successfully vest.

## Dependencies

This module uses the Cosmos SDK `x/auth` module and `x/auth/vesting` sub-module definitions of `Account` and `VestingAccount`. The actual state of a `ValidatorVestingAccount` is stored in the `x/auth` keeper, while this module merely stores a list of addresses that correspond to validator vesting accounts for fast iteration.
