<!--
order: 3
-->

# Messages

Users can submit various messages to the evmutil module which trigger state changes detailed below.

## MsgConvertERC20ToCoin

`MsgConvertCoinToERC20` converts a Kava ERC20 coin to sdk.Coin.

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

`MsgConvertCoinToERC20` converts sdk.Coin to Kava ERC20.

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

## MsgEVMCall

`MsgEVMCall` calls the EVM with data.

```protobuf
service Msg {
  // EVMCall defined a method for submitting an EVM call.
  rpc EVMCall(MsgEVMCall) returns (MsgEVMCallResponse);
}

// MsgEVMCall encapsulates an Ethereum call as a SDK message.
message MsgEVMCall {
  // Hex formatted address of the recipient.
  string to = 1;

  // Abi for the contract call, used for decoding contract call data.
  string fn_abi = 2;

  // Data payload of the call in hex string.
  string data = 3;

  // amount defines the integer value of the transaction amount.
  string amount = 4 [
    (cosmos_proto.scalar)  = "cosmos.Int",
    (gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Int",
    (gogoproto.nullable)   = false
  ];

  // authority is the address of the account that must be the signer.
  string authority = 5 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}
```

### State Changes

- EVM is called from the msg authority's 0x address with the provided data and amount. The msg authority must be the module's expected authority bech32 address. By default, this is the x/community module.
