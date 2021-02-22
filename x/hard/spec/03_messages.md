<!--
order: 3
-->

# Messages

There are three messages in the hard module. Deposit allows users to deposit assets to the hard module. In version 2, depositors will be able to use their deposits as collateral to borrow from hard. Withdraw removes assets from the hard module, returning them to the user. Claim allows users to claim earned HARD tokens.

```go
// MsgDeposit deposit collateral to the hard module.
type MsgDeposit struct {
  Depositor sdk.AccAddress `json:"depositor" yaml:"depositor"`
  Amount    sdk.Coins      `json:"amount" yaml:"amount"`
}
```

This message creates a `Deposit` object if one does not exist, or updates an existing one, as well as creating/updating the necessary indexes and synchronizing any outstanding interest. The `Amount` of coins is transferred from `Depositor` to the hard module account. The global variable for `TotalSupplied` is updated.

```go
// MsgWithdraw withdraw from the hard module.
type MsgWithdraw struct {
  Depositor sdk.AccAddress `json:"depositor" yaml:"depositor"`
  Amount    sdk.Coins      `json:"amount" yaml:"amount"`
}
```

This message decrements a `Deposit` object, or deletes one if the `Amount` specified is greater than or equal to the total deposited amount, as well as creating/updating the necessary indexes and synchronizing any outstanding interest. For example, a message which requests to withdraw 100xyz tokens, if `Depositor` has only deposited 50xyz tokens, will withdraw the full 50xyz tokens. The `Amount` of coins, or the current deposited amount, whichever is lower, is transferred from the hard module account to `Depositor`. The global variable for `TotalSupplied` is updated.

```go
// MsgBorrow borrows funds from the hard module.
type MsgBorrow struct {
  Borrower sdk.AccAddress `json:"borrower" yaml:"borrower"`
  Amount   sdk.Coins      `json:"amount" yaml:"amount"`
}
```

This message creates a `Borrow` object is one does not exist, or updates an existing one, as well as creating/updating the necessary indexes and synchronizing any outstanding interest. The `Amount` of coins is transferred from the hard module account to `Depositor`. The global variable for `TotalBorrowed` is updated.

```go
// MsgRepay repays funds to the hard module.
type MsgRepay struct {
	Sender sdk.AccAddress `json:"sender" yaml:"sender"`
	Owner  sdk.AccAddress `json:"owner" yaml:"owner"`
	Amount sdk.Coins      `json:"amount" yaml:"amount"`
}
```

This message decrements a `Borrow` object, or deletes one if the `Amount` specified is greater than or equal to the total borrowed amount, as well as creating/updating the necessary indexes and synchronizing any outstanding interest. For example, a message which requests to repay 100xyz tokens, if `Owner` has only deposited 50xyz tokens, the `Sender` will repay the full 50xyz tokens. The `Amount` of coins, or the current borrow amount, is transferred from `Sender`. The global variable for `TotalBorrowed` is updated.

```go
// MsgLiquidate attempts to liquidate a borrower's borrow
type MsgLiquidate struct {
  Keeper   sdk.AccAddress `json:"keeper" yaml:"keeper"`
  Borrower sdk.AccAddress `json:"borrower" yaml:"borrower"`
}
```

This message deletes `Borrower's` `Deposit` and `Borrow` objects if they are below the required LTV ratio. The keeper (the sender of the message) is rewarded a portion of the borrow position, according to the `KeeperReward` governance parameter. The coins from the `Deposit` are then sold at auction (see [auction module](../../auction/spec/README.md)), which any remaining tokens returned to `Borrower`. After being liquidated, `Borrower` no longer must repay the borrow amount. The global variables for `TotalSupplied` and `TotalBorrowed` are updated.
