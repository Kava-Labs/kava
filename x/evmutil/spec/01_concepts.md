<!--
order: 1
-->

# Concepts

## EVM Gas Denom

In order to use the EVM and be compatible with existing clients, the gas denom used by the EVM must be in 18 decimals. Since `ukava` has 6 decimals of precision, it cannot be used as the EVM gas denom directly.

Thus, in order to use the Kava token on the EVM, the `evmutil` module provides an `EvmBankKeeper` that enables the usage of `akava` on the EVM by using an account's `ukava` balance and its excess `akava` balance in the module store.

## `EvmBankKeeper` Overview

The `EvmBankKeeper` provides access to an account's **total** `akava` balance and the ability to transfer, mint, and burn `akava`. If anything other than the `akava` denom is requested, the `EvmBankKeeper` will panic.

```go
type BankKeeper interface {
	evmtypes.BankKeeper
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
}
```

The keeper implements the `x/evm` module's `BankKeeper` interface to enable the usage of `akava` denom on the EVM.

### `x/evm` Parameter

Since the EVM denom `akava` is required to use the `EvmBankKeeper`, it is necessary to set the `EVMDenom` param of the `x/evm` module to `akava`.

### `akava` Balance Calculation

The `akava` balance of an account is derived from an account's **spendable** `ukava` balance times 10^12 (to derive its `ukava` equivalent), plus the account's excess `akava` balance that can be accessed by the module `Keeper`.

### Conversions Between `akava` & `ukava`

When an account does not have sufficient `akava` to cover a transfer or burn, the `EvmBankKeeper` will try to swap 1 `ukava` to its equivalent `akava` amount. It does this by transferring 1 `ukava` from the sender to the `evmutil` module account, then adding the equivalent `akava` amount to the sender's balance via the keeper's `AddBalance`.

In reverse, if an account has enough `akava` balance for one or more `ukava`, the excess `akava` balance will be converted to `ukava`. This is done by removing the excess `akava` balance tracked by `evmutil` module store, then transferring the equivalent `ukava` coins from the `evmutil` module account to the target account.

The swap logic ensures that all `akava` is backed by the equivalent `ukava` balance stored in the module account.

## Module Keeper

The module Keeper provides access to an account's excess `akava` balance and the ability to update the balance.
