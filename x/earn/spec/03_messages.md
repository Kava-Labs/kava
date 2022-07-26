# Messages

## MsgDeposit

`MsgDeposit` adds supply to a specified vault:

```go
// MsgDeposit represents a message for depositing assedts into a vault
type MsgDeposit struct {
	// depositor represents the address to deposit funds from
	Depositor string `protobuf:"bytes,1,opt,name=depositor,proto3" json:"depositor,omitempty"`
	// Amount represents the token to deposit. The vault corresponds to the denom
	// of the amount coin.
	Amount types.Coin `protobuf:"bytes,2,opt,name=amount,proto3" json:"amount"`
}
```

The first deposit to a vault results in a `VaultRecord` being created. For each
deposit, a `VaultShareRecord` is created or updated, depending on if the
depositor has an existing deposit. The deposited tokens are converted to shares
which are equal to the amount of assets they deposited.

## MsgWithdraw

`MsgWithdraw` removes supply to a specified vault:

```go
// MsgWithdraw represents a message for withdrawing liquidity from a vault
type MsgWithdraw struct {
	// from represents the address we are withdrawing for
	From string `protobuf:"bytes,1,opt,name=from,proto3" json:"from,omitempty"`
	// Amount represents the token to withdraw. The vault corresponds to the denom
	// of the amount coin.
	Amount types.Coin `protobuf:"bytes,2,opt,name=amount,proto3" json:"amount"`
}
```

When withdrawing, the `VaultShareRecord` of the user will be decremented by the
corresponding amount of shares, or deleted in the case that all assets have
been withdrawn. If all shares of a vault have been withdrawn, the `VaultRecord`
will be deleted.
