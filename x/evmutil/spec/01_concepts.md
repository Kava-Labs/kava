<!--
order: 1
-->

# Concepts

## ERC20 token <> sdk.Coin Conversion

`x/evmutil` facilitates moving assets between Kava's EVM and Cosmos co-chains. This must be handled differently depending on which co-chain to which the asset it native. The messages controlling these flows involve two accounts:
1. The _initiator_ who sends coins from their co-chain
2. The _receiver_ who receives coins on the other co-chain

When converting assets from the EVM to the Cosmos co-chain, the initiator is an 0x EVM address and the receiver is a `kava1` Bech32 address.

When converting assets from the Cosmos co-chain to the EVM, the initiator is a `kava1` Bech32 address and the receiver is an 0x EVM address.

### Cosmos-Native Assets

`sdk.Coin`s native to the Cosmos co-chain can be converted to an ERC-20 representing the coin in the EVM. This works by transferring the coin from the initiator to `x/evmutil`'s module account and then minting an ERC-20 token to the receiver. Converting back, the initiator's ERC-20 representation of the coin is burned and the original Cosmos-native asset is transferred to the receiver.

Cosmos-native asset conversion is done through the use of the `MsgConvertCosmosCoinToERC20` & `MsgConvertCosmosCoinFromERC20` messages (see **[Messages](03_messages.md)**).

Only Cosmos co-chain denominations that are in the `AllowedCosmosDenoms` param (see **[Params](05_params.md)**) can be converted via these messages.

`AllowedCosmosDenoms` can be altered through governance.

The ERC20 contracts are deployed and managed by x/evmutil. The contract is deployed on first convert of the coin. Once deployed, the addresses of the contracts can be queried via the `DeployedCosmosCoinContracts` query (`deployed_cosmos_coin_contracts` endpoint).

If a denom is removed from the `AllowedCosmosDenoms` param, existing ERC20 tokens can be converted back to the underlying sdk.Coin via `MsgConvertCosmosCoinFromERC20`, but no conversions from sdk.Coin -> ERC via `MsgConvertCosmosCoinToERC20` are allowed.

### EVM-Native Assets

ERC-20 tokens native to the EVM can be converted into an `sdk.Coin` in the Cosmos ecosystem. This works by transferring the tokens to `x/evmutil`'s module account and then minting an `sdk.Coin` to the receiver. Converting back is the inverse: the `sdk.Coin` of the initiator is burned and the original ERC-20 tokens that were locked into the module account are transferred back to the receiver.

EVM-native asset conversion is done through the use of the `MsgConvertERC20ToCoin` & `MsgConvertCoinToERC20` messages (see **[Messages](03_messages.md)**).

Only ERC20 contract address that are in the `EnabledConversionPairs` param (see **[Params](05_params.md)**) can be converted via these messages.

`EnabledConversionPairs` can be altered through governance.
