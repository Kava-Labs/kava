# Messages

## CreateCDP

CreateCDP makes a new CDP in state, adding collateral from the sender, and drawing `Principle` debt.

```go
type MsgCreateCDP struct {
	Sender     sdk.AccAddress
	Collateral sdk.Coins
	Principal  sdk.Coins
}
```

## Deposit

Deposit adds collateral to an existing cdp.

```go
type MsgDeposit struct {
	Owner      sdk.AccAddress
	Depositor  sdk.AccAddress
	Collateral sdk.Coins
}
```

## Withdraw

Withdraw removes collateral from a CDP. Collateral cannot be removed if it would put the CDP under the collateralization ratio.

```go
type MsgWithdraw struct {
	Owner      sdk.AccAddress
	Depositor  sdk.AccAddress
	Collateral sdk.Coins
}
```

## DrawDebt

DrawDebt creates debt in a CDP, minting new pegged asset which is sent to the sender.

```go
type MsgDrawDebt struct {
	Sender    sdk.AccAddress
	CdpDenom  string
	Principal sdk.Coins
}
```

## RepayDebt

RepayDebt removes some debt from a CDP and burns the corresponding amount of pegged asset from the sender.

```go
type MsgRepayDebt struct {
	Sender   sdk.AccAddress
	CdpDenom string
	Payment  sdk.Coins
}
```
