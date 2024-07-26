<!--
order: 2
-->

# State

## Parameters and Genesis State

`Parameters` define the list of conversion pairs allowed to be converted between Kava ERC20 tokens & sdk.Coins, and the list of native cosmos sdk.Coins that are allowed to be converted to ERC20s.

```protobuf
// Params defines the evmutil module params
message Params {
  // enabled_conversion_pairs defines the list of conversion pairs allowed to be
  // converted between Kava ERC20 and sdk.Coin
  repeated ConversionPair enabled_conversion_pairs = 4;

  // allowed_cosmos_denoms is a list of denom & erc20 token metadata pairs.
  // if a denom is in the list, it is allowed to be converted to an erc20 in the evm.
  repeated AllowedCosmosCoinERC20Token allowed_cosmos_denoms = 1;
}

// ConversionPair defines a Kava ERC20 address and corresponding denom that is
// allowed to be converted between ERC20 and sdk.Coin
message ConversionPair {
  // ERC20 address of the token on the Kava EVM
  bytes kava_erc20_address = 1;
  // Denom of the corresponding sdk.Coin
  string denom = 2;
}

// AllowedCosmosCoinERC20Token defines allowed cosmos-sdk denom & metadata
// for evm token representations of sdk assets.
// NOTE: once evm token contracts are deployed, changes to metadata for a given
// cosmos_denom will not change metadata of deployed contract.
message AllowedCosmosCoinERC20Token {
  option (gogoproto.goproto_getters) = false;

  // Denom of the sdk.Coin
  string cosmos_denom = 1;
  // Name of ERC20 contract
  string name = 2;
  // Symbol of ERC20 contract
  string symbol = 3;
  // Number of decimals ERC20 contract is deployed with.
  uint32 decimals = 4;
}
```

`GenesisState` defines the state that must be persisted when the blockchain stops/restarts in order for normal function of the evmutil module to resume.

```protobuf
message GenesisState {
  // previously stored accounts containing fractional balances.
  reserved 1;
  Params params = 2 [(gogoproto.nullable) = false];
}
```

## Deployed Cosmos Coin Contract Addresses

Addresses for the ERC20 contracts representing cosmos-sdk `Coin`s are kept in the module store. They are stored as bytes by the cosmos-sdk denom they represent.

Example:
If a contract for representing the cosmos-sdk denom `cow` as an ERC20 in the EVM is deployed by the module to the address `0xbeef00000000000000000000000000000000beef`, the module store will contain:

`0x01 | bytes("cow") => bytes(0xbeef00000000000000000000000000000000beef)`

Where `0x01` is the `DeployedCosmosCoinContractKeyPrefix` defined in [keys.go](../types/keys.go).

## Store

For complete implementation details for how items are stored, see [keys.go](../types/keys.go). `x/evmutil` store state consists of accounts and deployed contract addresses.
