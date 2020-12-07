<!--
order: 2
-->

# State

For detail on the state tracked by the cdp module see the types package. In particular [keys.go](../types/keys.go) describes how state is stored in the key-value store.

## Module Accounts

The cdp module account controls two module accounts:

**CDP Account:** Stores the deposited cdp collateral, and the debt coins for the debt in all the cdps.

**Liquidator Account:** Stores debt coins that have been seized by the system, and any stable asset that has been raised through auctions.

## CDP

A CDP is a struct representing a debt position owned by one address. It has one collateral type and records the debt that has been drawn and how much fees should be repaid.

Only an owner is authorized to draw or repay debt, but anyone can deposit collateral to a CDP. Deposits are scoped per address and are recorded separately in `Deposit` types. Depositors are free to withdraw their collateral provided it does not put the CDP below the liquidation ratio.

The CDP's collateral always equal to the total of the deposits.

```go
type CDP struct {
    ID              uint64
    Owner           sdk.AccAddress
    Type            string
    Collateral      sdk.Coin
    Principal       sdk.Coin
    AccumulatedFees sdk.Coin
    FeesUpdated     time.Time
    InterestFactor  sdk.Dec
}
```

CDPs are stored with three database indexes for faster lookup:

- by collateral ratio - to look up cdps that are close to the liquidation ratio
- by collateral denom - to look up cdps with a particular collateral asset
- by owner index - to look up cdps that an address is the owner of

## Deposit

A Deposit is a struct recording collateral added to a CDP by one address. The address only has authorization to change their deposited amount (provided it does not put the CDP below the liquidation ratio).

```go
type Deposit struct {
    CdpID         uint64
    Depositor     sdk.AccAddress
    Amount        sdk.Coin
}
```

## Params

Module parameters controlled by governance. See [Parameters](04_params.md) for details.

## NextCDPID

A global counter used to create unique CDP ids.

## DebtDenom

The name of the internal debt coin. Its value can be configured at genesis.

## GovDenom

The name of the internal governance coin. Its value can be configured at genesis.

## Total Principle

Sum of all non seized debt plus accumulated fees.

## Previous Savings Distribution Time

A record of the last block time when the savings rate was distributed
