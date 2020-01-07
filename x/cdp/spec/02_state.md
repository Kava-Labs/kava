# State

Listed here is the state stored by the module.
For details see the types package. In particular `keys.go` describes how state is stored in the key-value store.
<!--
Store key structures are not listed here - they seem like an implementation detail best documented by code comments.
-->

## CDP

A CDP is a struct representing a debt position owned by one address. It has one collateral type and records the debt that has been drawn and how much fees should be repaid.

Only an owner is authorized to draw or repay debt. But anyone can deposit to any CDP, and withdraw their debt (provided it does not put the CDP below the liquidation ratio).

The CDP's collateral must always equal the total of the deposits. It is stored in the CDP for efficiency purposes.

```go
type CDP struct {
    ID              uint64
    Owner           sdk.AccAddress
    Collateral      sdk.Coins
    Principal       sdk.Coins
    AccumulatedFees sdk.Coins
    FeesUpdated     time.Time
}
```

These are stored with a couple of database indexes for faster lookup:

- by collateral ratio - to look up cdps that are close to the liquidation ratio
- by owner index - to look up cdps that an address has deposited to

## Deposit

A Deposit is a struct recording collateral added to a CDP by one address. The address only has authorization to change their deposited amount.

The liquidation flag indicates whether this deposit has been seized for sell off.
<!-- TODO when is the flag unset? could the deposit be removed on liquiation to make things simpler? -->
<!-- TODO can Amount be more than one coin? suggest change to sdk.Coin if not -->

```go
type Deposit struct {
    CdpID         uint64
    Depositor     sdk.AccAddress
    Amount        sdk.Coins
    InLiquidation bool
}
```

## Params

Module parameters controlled by governance. See [Parameters](07_params.md) for details.

## NextCDPID

A global counter used to create unique CDP ids.

## DebtDenom

The name for the internal debt coin. It's set in the store rather than hard-coded so its value can be configured at genesis.

## Total Principle

Total pegged assets drawn for each pegged asset denom.
<!-- TODO is this different from the total amount of the assets created? ie does it include amount in liquidator? -->

## FeeRate

<!-- TODO -->

## Previous Block Time

A record of the last block time used in calculating fees.
<!-- TODO add details -->
