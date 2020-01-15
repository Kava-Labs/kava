# Concepts

## Collateralized Debt Positions

CDPs enable the creation of a stable pegged asset by collateralization with another on chain asset.

A CDP is scoped to one collateral type. It has one primary owner, and a set of "depositors". The depositors can deposit and withdraw collateral to the CDP. The owner can draw pegged assets (creating debt) and repay them to cancel the debt.

User interactions with this module:

- create a new cdp by depositing some type of coin as collateral
- withdraw newly minted stable coin from this CDP (up to a fraction of the value of the collateral)
- repay debt by paying back stable coins (including paying any fees accrued)
- remove collateral and close CDP

Automatic actions:

- CDPs that fall below the liquidation ratio (how over-collateralised the debt is) are seized and collateral is auctioned off through another auction module.
- Seized debt is netted with the proceeds from auction sales. Any remaining is rebalanced by triggering auctions.

An up to date price of each collateral is required, and is provided by a "pricefeed" module.

Although a CDP is restricted to one type of collateral, users can create other CDPs that store different types. The module parameters store allowed collateral types.

Pegged assets are pegged to one external price, however they can have multiple denominations within the system.
<!-- TODO what parts treat different denoms as interchangeable and what parts do not -->

## Fees

When a user repays stable coin withdrawn from a CDP, they must also pay an additional amount known as the fee.

This is calculated according to the amount of stable coin withdrawn and the time withdrawn for, much like interest on a loan.

Fees create incentives to open or close CDPs and can be changed by governance to help keep the system functioning through changing market conditions.

<!-- TODO how they are calculated -->

## Dependency: supply

The CDP module has a 'Module Account' which store user's assets. It relies on a supply keeper to move assets between it's module account and other user and module accounts.

## Dependency: pricefeed

The CDP module needs to know the current real world price of collateral assets in order to calculate if CDPs are under collateralized. It calls another module to return a price for a given collateral in units of the pegged asset (usually US Dollars).

CDP assumes the price feed is accurate, leaving it to external modules to achieve this.
