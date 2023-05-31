<!--
order: 3
-->

# Messages

Users can submit various messages to the evmutil module which trigger state changes detailed below.

## MsgConvertCosmosCoinToERC20

`MsgConvertCosmosCoinToERC20` converts an sdk.Coin to an ERC20. This message is for moving Cosmos-native assets from the Cosmos ecosystem to the EVM.

Upon first conversion, the message also deploys the ERC20 contract that will represent the cosmos-sdk asset in the EVM. The contract is owned by the `x/evmutil` module.

```proto
service Msg {
  // ConvertCosmosCoinToERC20 defines a method for converting a cosmos sdk.Coin to an ERC20.
  rpc ConvertCosmosCoinToERC20(MsgConvertCosmosCoinToERC20) returns (MsgConvertCosmosCoinToERC20Response);
}

// MsgConvertCosmosCoinToERC20 defines a conversion from cosmos sdk.Coin to ERC20.
message MsgConvertCosmosCoinToERC20 {
  // Kava bech32 address initiating the conversion.
  string initiator = 1;
  // EVM hex address that will receive the ERC20 tokens.
  string receiver = 2;
  // Amount is the sdk.Coin amount to convert.
  cosmos.base.v1beta1.Coin amount = 3;
}
```

### State Changes

- The `AllowedCosmosDenoms` param from `x/evmutil` is checked to ensure the conversion is allowed.
- The module's store is checked for the address of the deployed ERC20 contract. If none is found, a new contract is deployed and its address is saved to the module store.
- The `amount` is deducted from the `initiator`'s balance and transferred to the module account.
- An equivalent amount of ERC20 tokens are minted by `x/evmutil` to the `receiver`.

## MsgConvertCosmosCoinFromERC20

`MsgConvertCosmosCoinFromERC20` is the inverse of `MsgConvertCosmosCoinToERC20`. It converts an ERC20 representation of a cosmos-sdk coin back to its underlying sdk.Coin.


```proto
service Msg {
  // ConvertCosmosCoinFromERC20 defines a method for converting a cosmos sdk.Coin to an ERC20.
  rpc ConvertCosmosCoinFromERC20(MsgConvertCosmosCoinFromERC20) returns (MsgConvertCosmosCoinFromERC20Response);
}

// MsgConvertCosmosCoinFromERC20 defines a conversion from ERC20 to cosmos coins for cosmos-native assets.
message MsgConvertCosmosCoinFromERC20 {
  // EVM hex address initiating the conversion.
  string initiator = 1;
  // Kava bech32 address that will receive the cosmos coins.
  string receiver = 2;
  // Amount is the amount to convert, expressed as a Cosmos coin.
  cosmos.base.v1beta1.Coin amount = 3;
}
```

### State Changes

- The `amount` is transferred from the `x/evmutil` module account to the `receiver`.
- The same amount of the corresponding ERC20 is burned from the `initiator` account in the EVM.

## MsgConvertERC20ToCoin

`MsgConvertCoinToERC20` converts a Kava ERC20 coin to sdk.Coin. This message is for moving EVM-native assets from the EVM to the Cosmos ecosystem.

```protobuf
service Msg {
  // ConvertERC20ToCoin defines a method for converting Kava ERC20 to sdk.Coin.
  rpc ConvertERC20ToCoin(MsgConvertERC20ToCoin) returns (MsgConvertERC20ToCoinResponse);
}

// MsgConvertERC20ToCoin defines a conversion from Kava ERC20 to sdk.Coin.
message MsgConvertERC20ToCoin {
  // EVM 0x hex address initiating the conversion.
  string initiator = 1;
  // Kava bech32 address that will receive the converted sdk.Coin.
  string receiver = 2;
  // EVM 0x hex address of the ERC20 contract.
  string kava_erc20_address = 3;
  // ERC20 token amount to convert.
  string amount = 4 [
    (gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Int",
  ];
}
```

### State Changes

- The `EnabledConversionPairs` param from `x/evmutil` is checked to ensure the conversion pair is enabled.
- The initiator's ERC20 token from `kava_erc20_address` is locked by transferring it from the initiator's 0x address to the `x/evmutil` module account's 0x address.
- The same amount of sdk.Coin are minted for the corresponding denom of the `kava_erc20_address` in the `EnabledConversionPairs` param. The coins are then transferred to the receiver's Kava address.

## MsgConvertCoinToERC20

`MsgConvertCoinToERC20` converts sdk.Coin to Kava ERC20. This message is for moving EVM-native assets from the Cosmos ecosystem back to the EVM.

```protobuf
service Msg {
  // ConvertCoinToERC20 defines a method for converting sdk.Coin to Kava ERC20.
  rpc ConvertCoinToERC20(MsgConvertCoinToERC20) returns (MsgConvertCoinToERC20Response);
}

// MsgConvertCoinToERC20 defines a conversion from sdk.Coin to Kava ERC20.
message MsgConvertCoinToERC20 {
  // Kava bech32 address initiating the conversion.
  string initiator = 1;
  // EVM 0x hex address that will receive the converted Kava ERC20 tokens.
  string receiver = 2;
  // Amount is the sdk.Coin amount to convert.
  cosmos.base.v1beta1.Coin amount = 3;
}
```

### State Changes

- The `EnabledConversionPairs` param from `x/evmutil` is checked to ensure the conversion pair is enabled.
- The specified sdk.Coin is moved from the initiator's address to the module account and burned.
- The same amount of ERC20 coins are sent from the `x/evmutil` module account to the 0x receiver address.
