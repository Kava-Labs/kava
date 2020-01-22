# Messages

Users can submit various messages to the cdp module which trigger state changes detailed below.

## CreateCDP

CreateCDP sets up and stores a new CDP, adding collateral from the sender, and drawing `Principle` debt.

```go
type MsgCreateCDP struct {
    Sender     sdk.AccAddress
    Collateral sdk.Coins
    Principal  sdk.Coins
}
```

State changes:

- a new CDP is created, `Sender` becomes CDP owner
- collateral taken from `Sender` and sent to cdp module account, new `Deposit` created
- `Principal` stable coins are minted and sent to `Sender`
- equal amount of internal debt coins created and stored in cdp module account

## Deposit

Deposit adds collateral to a CDP in the form of a deposit. Collateral is taken from `Depositor`.

```go
type MsgDeposit struct {
    Owner      sdk.AccAddress
    Depositor  sdk.AccAddress
    Collateral sdk.Coins
}
```

State Changes:

- `Collateral` taken from depositor and sent to cdp module account
- the depositor's `Deposit` struct is updated or a new one created
- cdp fees are updated (see below)

## Withdraw

Withdraw removes collateral from a CDP, provided it would not put the CDP under the liquidation ratio. Collateral is removed from one deposit only.

```go
type MsgWithdraw struct {
    Owner      sdk.AccAddress
    Depositor  sdk.AccAddress
    Collateral sdk.Coins
}
```

State Changes:

- `Collateral` coins are sent from the cdp module account to `Depositor`
- `Collateral` amount of coins subtracted from the `Deposit` struct
- cdp fees are updated (see below)

## DrawDebt

DrawDebt creates debt in a CDP, minting new stable asset which is sent to the sender.

```go
type MsgDrawDebt struct {
    Sender    sdk.AccAddress
    CdpDenom  string
    Principal sdk.Coins
}
```

State Changes:

- mint `Principal` coins and send them to `Sender`, updating the CDP's `Principal` field
- mint equal amount of internal debt coins and store in the module account
- increment total principal for principal denom
- cdp fees are updated (see below)

## RepayDebt

RepayDebt removes some debt from a CDP and burns the corresponding amount of stable asset from the sender. If all debt is repaid, the collateral is returned to depositors and the cdp is removed from the store

```go
type MsgRepayDebt struct {
    Sender   sdk.AccAddress
    CdpDenom string
    Payment  sdk.Coins
}
```

State Changes:

- burn `Payment` coins taken from `Sender`, updating the CDP by reducing `Principal` field by `Paymment`
- burn an equal amount of internal debt coins
- decrement total principal for payment denom
- cdp fees are updated (see below)
- if fees and principal are zero, return collateral to depositors:
  - For each deposit, send coins from the cdp module account to the depositor, and delete the deposit struct from store.

## Fees

When CDPs are updated by the above messages the fees accumulated since the last update are calculated and added on.

```
feesAccumulated = (outstandingDebt * (feeRate^periods)) - outstandingDebt
```

where:

- `outstandingDebt` is the CDP's `Principal` plus `AccumulatedFees`
- `periods` is the number of seconds since last fee update
- `feeRate` is the per second debt interest rate

## Database Indexes

When CDPs are update by the above messages the database indexes are also updated.