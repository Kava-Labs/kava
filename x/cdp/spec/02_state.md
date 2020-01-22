# State

Listed here is the state stored by the module.
For details see the types package. In particular `keys.go` describes how state is stored in the key-value store.
<!--
Store key structures are not listed here - they seem like an implementation detail best documented by code comments.
-->

## Module Accounts

The cdp module account controls two module accounts:

**CDP Account:** Stores the deposited cdp collateral, and the debt coins for the debt in all the cdps.

**Liquidator Account:** Stores debt coins that have been seized by the system, and pegged asset that has been raised through auctions.

## CDP

A CDP is a struct representing a debt position owned by one address. It has one collateral type and records the debt that has been drawn and how much fees should be repaid.

Only an owner is authorized to draw or repay debt. But anyone can deposit or withdraw collateral to a CDP (provided it does not put the CDP below the liquidation ratio). Deposits are recorded separately in `Deposit` types.

The CDP's collateral always equal the total of the deposits.

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

CDPs are stored with a couple of database indexes for faster lookup:

- by collateral ratio - to look up cdps that are close to the liquidation ratio
- by owner index - to look up cdps that an address has deposited to

## Deposit

A Deposit is a struct recording collateral added to a CDP by one address. The address only has authorization to change their deposited amount (provided it does not put the CDP below the liquidation ratio).

```go
type Deposit struct {
    CdpID         uint64
    Depositor     sdk.AccAddress
    Amount        sdk.Coins
}
```

## Params

Module parameters controlled by governance. See [Parameters](07_params.md) for details.

## NextCDPID

A global counter used to create unique CDP ids.

## DebtDenom

The name for the internal debt coin. Its value can be configured at genesis.

## Total Principle

Total pegged assets minted plus accumulated fees. This aggregate of all debt is used to calculate the new debt created every block due to due to fees.

## Previous Block Time

A record of the last block time used to calculate fees.
