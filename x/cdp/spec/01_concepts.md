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

- CDPs that fall below the liquidation ratio (how over-collateralized the debt is) are seized and collateral is auctioned off through another auction module.
- Seized debt is netted with the proceeds from auction sales. Any remaining is rebalanced by triggering auctions.

An up-to-date price of each collateral is required, and is provided by a "pricefeed" module.

Although a CDP is restricted to one type of collateral asset, users can create other CDPs collateralized by different assets. Allowed types can be updated by governance.

Pegged assets are pegged to one external price, however they can have multiple denominations within the system.

## Liquidation & Stability System

In the event of a decrease in the price of the collateral, the total value of all collateral in CDPs may drop below the value of all the issued pegged asset. This undesirable event is countered through two mechanisms.

**CDP Liquidations** Each CDP is monitored for ratio of collateral value to debt value. When this drops too low the collateral and debt is seized by the system. The collateral is sold off through an auction to bring in pegged asset which is burned against the seized debt.

**Debt Auctions** In extreme cases where liquidations fail to raise enough to cover the seized debt, another mechanism kicks in: Debt Auctions. System governance token is minted and sold through auction to raise enough pegged asset to cover the remaining debt. The governors of the system represent the lenders of last resort.

## Fees

When a user repays stable coin withdrawn from a CDP, they must also pay an additional amount known as the fee.

This is calculated according to the amount of stable coin withdrawn and the time withdrawn for. Like interest on a loan fees grow at a compounding percentage of original debt.

Fees create incentives to open or close CDPs and can be changed by governance to help keep the system functioning through changing market conditions.

A further fee is applied on liquidation of a CDP. Normally when the collateral is sold to cover the debt, any excess not sold is returned to the CDP holder. The liquidation fee reduces the amount of excess collateral returned, representing a cut that the system takes.

Fees accumulate to the system before being sold at auction for governance token. These are then burned, acting as incentive for safe governance of the system.

## Dependency: supply

The CDP module has two 'Module Accounts' which store user's assets. It relies on a supply keeper to move assets between it's module account and other user and module accounts.

## Dependency: pricefeed

The CDP module needs to know the current price of collateral assets in order to determine if CDPs are under collateralized. It calls a "pricefeed" module to return a price for a given collateral in units of the pegged asset (usually US Dollars).
