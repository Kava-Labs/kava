<!--
order: 2
-->

# State

## Parameters and Genesis State

`Parameters` define the list of conversion pairs allowed to be converted between Kava ERC20 tokens & sdk.Coins, and the list of native sdk.Coins that are allowed to be converted to ERC20s.

```protobuf
// Params defines the evmutil module params
message Params {
  // enabled_conversion_pairs defines the list of conversion pairs allowed to be
  // converted between Kava ERC20 and sdk.Coin
  repeated ConversionPair enabled_conversion_pairs = 4;

  // allowed_native_denoms is a list of denom & erc20 token metadata pairs.
  // if a denom is in the list, it is allowed to be converted to an erc20 in the evm.
  repeated AllowedNativeCoinERC20Token allowed_native_denoms = 1;
}

// ConversionPair defines a Kava ERC20 address and corresponding denom that is
// allowed to be converted between ERC20 and sdk.Coin
message ConversionPair {
  // ERC20 address of the token on the Kava EVM
  bytes kava_erc20_address = 1;
  // Denom of the corresponding sdk.Coin
  string denom = 2;
}

// AllowedNativeCoinERC20Token defines allowed sdk denom & metadata
// for evm token representations of sdk assets.
// NOTE: once evm token contracts are deployed, changes to metadata for a given
// sdk_denom will not change metadata of deployed contract.
message AllowedNativeCoinERC20Token {
  option (gogoproto.goproto_getters) = false;

  // Denom of the sdk.Coin
  string sdk_denom = 1;
  // Name of ERC20 contract
  string name = 2;
  // Symbol of ERC20 contract
  string symbol = 3;
  // Number of decimals ERC20 contract is deployed with.
  uint32 decimal = 4;
}

```

`GenesisState` defines the state that must be persisted when the blockchain stops/restarts in order for normal function of the evmutil module to resume.

```protobuf
message GenesisState {
  repeated Account accounts = 1 [(gogoproto.nullable) = false];
  Params params = 2 [(gogoproto.nullable) = false];
}
```

## Account

An `Account` is a struct representing the excess `akava` balance of an address.

Since an address's total `akava` balance is derived from its `ukava` balance and the excess `akava` balance stored by the `Account` struct, the `akava` balance here should never exceed 1 `ukava` (10^12 `akava`).

```protobuf
message Account {
  bytes address = 1;
  string balance = 2;
}
```

## Store

For complete implementation details for how items are stored, see [keys.go](../types/keys.go). `x/evmutil` store state consists of accounts.
