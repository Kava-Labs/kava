# Messages

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

- new CDP created, Sender becomes CDP owner
- collateral taken from Sender, sent to cdp module account, new Deposit created
- principal coins minted and sent to Sender
- equal amount of internal debt coins created and stored in cdp module account

## Deposit

Deposit adds collateral to a CDP in the form of a deposit. Collateral is taken from Depositor.

```go
type MsgDeposit struct {
    Owner      sdk.AccAddress
    Depositor  sdk.AccAddress
    Collateral sdk.Coins
}
```

State Changes:

- Collateral taken from depositor and sent to cdp module account.
- The depositor's Deposit struct is updated or a new one created.
- something to do with fees <!-- TODO -->

## Withdraw

Withdraw removes collateral from a CDP. Collateral cannot be removed if it would put the CDP under the collateralization ratio. Collateral is removed from one deposit only. Collateral is sent to Depositor.

```go
type MsgWithdraw struct {
    Owner      sdk.AccAddress
    Depositor  sdk.AccAddress
    Collateral sdk.Coins
}
```

State Changes:

- Something to do with fees <!-- TODO -->
- Collateral coins are sent from cdp's module account to Depositor.
- Collateral amount of coins subtracted from the Deposit struct. <!-- TODO should this delete deposit if empty?-->

## DrawDebt

DrawDebt creates debt in a CDP, minting new pegged asset which is sent to the sender.
<!-- TODO Can the sender own have same collateral multiple CDPs? if so how do they choose between them.  -->

```go
type MsgDrawDebt struct {
    Sender    sdk.AccAddress
    CdpDenom  string
    Principal sdk.Coins
}
```

State Changes:

- Mint Principal coins and send them to Sender.
- Mint equal amount of internal debt coins and store in cdp's module account.
- Update CDP struct with new principal and fees. <!-- TODO how fees are calculated -->
- Increment total principal for principal denom.

## RepayDebt

RepayDebt removes some debt from a CDP and burns the corresponding amount of pegged asset from the sender. If all debt is repaid, the collateral is returned to depositors and the cdp is removed from the store

```go
type MsgRepayDebt struct {
    Sender   sdk.AccAddress
    CdpDenom string
    Payment  sdk.Coins
}
```

State Changes:

- Burn Payment coins taken from Sender.
- Burn an equal amount of internal debt coins.
- Update CDP by reducing Principal by Paymment, also update fees <!-- TODO -->
- Decrement total principal for payment denom.
- If fees and principal are zero, return collateral:
  - For each deposit, send coins from the cdp's module account to the depositor, and delete the deposit struct from store.
