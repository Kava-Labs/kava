# State

## State Objects

The `x/earn` module keeps the following in state:

| State Object        | Description                 | Key                              | Value                      |
| ------------------- | --------------------------- | -------------------------------- | -------------------------- |
| Vault Records       | List of vault records       | `[]byte{1} + []byte(vaultDenom)` | `[]byte{VaultRecord}`      |
| Vault Share Records | List of vault share records | `[]byte{2} + []byte(accAddress)` | `[]byte{VaultShareRecord}` |

## VaultRecord

Vault records track the total supply of a vault in state.

```go
// VaultRecord is the state of a vault and is used to store the state of a
// vault.
type VaultRecord struct {
	// Denom is the only supported denomination of the vault for deposits and
	// withdrawals.
	Denom string `protobuf:"bytes,1,opt,name=denom,proto3" json:"denom,omitempty"`
	// TotalSupply is the total supply of the vault, denominated **only** in the
	// user deposit/withdrawal denom, must be the same as the Denom field.
	TotalSupply types.Coin `protobuf:"bytes,2,opt,name=total_supply,json=totalSupply,proto3" json:"total_supply"`
}
```

## VaultShareRecord

Vault share records track the total amount of shares an account has for all
vaults.

```go
// VaultShareRecord defines the vault shares owned by a depositor.
type VaultShareRecord struct {
	// Depositor represents the owner of the shares
	Depositor github_com_cosmos_cosmos_sdk_types.AccAddress `protobuf:"bytes,1,opt,name=depositor,proto3,casttype=github.com/cosmos/cosmos-sdk/types.AccAddress" json:"depositor,omitempty"`
	// AmountSupplied represents the total amount a depositor has supplied to the
	// vault. The vault is determined by the coin denom.
	AmountSupplied github_com_cosmos_cosmos_sdk_types.Coins `protobuf:"bytes,2,rep,name=amount_supplied,json=amountSupplied,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"amount_supplied"`
}

```

## Genesis State

The `GenesisState` defines the state that must be persisted when the blockchain
stops/restarts in order for normal function of the bridge module to resume.

```go
// GenesisState defines the earn module's genesis state.
type GenesisState struct {
	// params defines all the paramaters related to earn
	Params Params `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
	// vault_records defines the available vaults
	VaultRecords VaultRecords `protobuf:"bytes,2,rep,name=vault_records,json=vaultRecords,proto3,castrepeated=VaultRecords" json:"vault_records"`
	// share_records defines the owned shares of each vault
	VaultShareRecords VaultShareRecords `protobuf:"bytes,3,rep,name=vault_share_records,json=vaultShareRecords,proto3,castrepeated=VaultShareRecords" json:"vault_share_records"`
}
```
