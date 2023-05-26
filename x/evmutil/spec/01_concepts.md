<!--
order: 1
-->

# Concepts

## EVM Gas Denom

In order to use the EVM and be compatible with existing clients, the gas denom used by the EVM must be in 18 decimals. Since `ukava` has 6 decimals of precision, it cannot be used as the EVM gas denom directly.

To use the Kava token on the EVM, the evmutil module provides an `EvmBankKeeper` that is responsible for the conversion of `ukava` and `akava`. A user's excess `akava` balance is stored in the `x/evmutil` store, while its `ukava` balance remains in the cosmos-sdk `x/bank` module.

## `EvmBankKeeper` Overview

The `EvmBankKeeper` provides access to an account's total `akava` balance and the ability to transfer, mint, and burn `akava`. If anything other than the `akava` denom is requested, the `EvmBankKeeper` will panic.

This keeper implements the `x/evm` module's `BankKeeper` interface to enable the usage of `akava` denom on the EVM.

### `x/evm` Parameter Requirements

Since the EVM denom `akava` is required to use the `EvmBankKeeper`, it is necessary to set the `EVMDenom` param of the `x/evm` module to `akava`.

### Balance Calculation of `akava`

The `akava` balance of an account is derived from an account's **spendable** `ukava` balance times 10^12 (to derive its `akava` equivalent), plus the account's excess `akava` balance that can be accessed via the module `Keeper`.

### `akava` <> `ukava` Conversion

When an account does not have sufficient `akava` to cover a transfer or burn, the `EvmBankKeeper` will try to swap 1 `ukava` to its equivalent `akava` amount. It does this by transferring 1 `ukava` from the sender to the `x/evmutil` module account, then adding the equivalent `akava` amount to the sender's balance in the module state.

In reverse, if an account has enough `akava` balance for one or more `ukava`, the excess `akava` balance will be converted to `ukava`. This is done by removing the excess `akava` balance in the module store, then transferring the equivalent `ukava` coins from the `x/evmutil` module account to the target account.

The swap logic ensures that all `akava` is backed by the equivalent `ukava` balance stored in the module account.

## ERC20 token <> sdk.Coin Conversion

`x/evmutil` facilitates moving assets between Kava's EVM and Cosmos co-chains. This must be handled differently depending on which co-chain to which the asset it native. The messages controlling these flows involve two accounts:
1. The _initiator_ who sends coins from their co-chain
2. The _receiver_ who receives coins on the other co-chain

When converting assets from the EVM to the Cosmos co-chain, the initiator is an 0x EVM address and the receiver is a `kava1` Bech32 address.

When converting assets from the Cosmos co-chain to the EVM, the initiator is a `kava1` Bech32 address and the receiver is an 0x EVM address.

### Cosmos-Native Assets

`sdk.Coin`s native to the Cosmos co-chain can be converted to an ERC-20 representing the coin in the EVM. This works by transferring the coin from the initiator to `x/evmutil`'s module account and then minting an ERC-20 token to the receiver. Converting back, the initiator's ERC-20 representation of the coin is burned and the original Cosmos-native asset is transferred to the receiver.

Cosmos-native asset converstion is done through the use of the `MsgConvertCosmosCoinToERC20` & `MsgConvertCosmosCoinFromERC20` messages (see **[Messages](03_messages.md)**).

Only Cosmos co-chain denominations that are in the `AllowedCosmosDenoms` param (see **[Params](05_params.md)**) can be converted via these messages.

`AllowedCosmosDenoms` can be altered through governance.

The ERC20 contracts are deployed and managed by x/evmutil. The contract is deployed on first convert of the coin. Once deployed, the addresses of the contracts can be queried via the `DeployedCosmosCoinContracts` query (`deployed_cosmos_coin_contracts` endpoint).

### EVM-Native Assets

ERC-20 tokens native to the EVM can be converted into an `sdk.Coin` in the Cosmos ecosystem. This works by transferring the tokens to `x/evmutil`'s module account and then minting an `sdk.Coin` to the receiver. Converting back is the inverse: the `sdk.Coin` of the initiator is burned and the original ERC-20 tokens that were locked into the module account are transferred back to the receiver.

EVM-native asset conversion is done through the use of the `MsgConvertERC20ToCoin` & `MsgConvertCoinToERC20` messages (see **[Messages](03_messages.md)**).

Only ERC20 contract address that are in the `EnabledConversionPairs` param (see **[Params](05_params.md)**) can be converted via these messages.

`EnabledConversionPairs` can be altered through governance.

## Module Keeper

The module Keeper provides access to an account's excess `akava` balance and the ability to update the balance.
