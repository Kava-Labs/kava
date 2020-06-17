<!--
order: 1
-->

# Concepts

## Collateralized Debt Positions

CDPs enable the creation of a stable asset by collateralization with another on chain asset.

A CDP is scoped to one collateral type. It has one primary owner, and a set of "depositors". The depositors can deposit and withdraw collateral to the CDP. The owner can draw stable assets (creating debt), deposit and withdraw collateral, and repay stable assets to cancel the debt.

Once created, stable assets are free to be transferred between users, but a CDP owner must repay their debt to get their collateral back.

User interactions with this module:

- create a new CDP by depositing a supported coin as collateral and minting debt
- deposit to a CDP controlled by a different owner address
- withdraw deposited collateral, if it doesn't put the CDP below the liquidation ratio
- issue stable coins from this CDP (up to a fraction of the value of the collateral)
- repay debt by paying back stable coins (including paying any fees accrued)
- remove collateral and close CDP

Module interactions:

- fees for all CDPs are updated each block
- the value of fees (surplus) is divded between users, via the savings rate, and owners of the governance token, via burning governance tokens proportional to surplus
- the value of an asset that is supported for CDPs is determined by querying an external pricefeed
- if the price of an asset puts a CDP below the liquidation ratio, the CDP is liquidated
- liquidated collateral is divided into lots and sent to an external auction module
- collateral that is returned from the auction module is returned to the account that deposited that collateral
- if auctions do not recover the desired amount of debt, debt auctions are triggered after a certain threshold of global debt is reached
- surplus auctions are triggered after a certain threshold of surplus is triggered

## Liquidation & Stability System

In the event of a decrease in the price of the collateral, the total value of all collateral in CDPs may drop below the value of all the issued stable assets. This undesirable event is countered through two mechanisms:

**CDP Liquidations** The ratio of collateral value to debt value in each CDP is monitored. When this drops too low the collateral and debt is automatically seized by the system. The collateral is sold off through an auction to bring in stable asset which is burned against the seized debt. The price used to determine liquidation is controlled by the `LiquidationMarketID` parameter, which can be the same as the `SpotMarketID` or use a different calculation of price, such as a time-weighted average.

**Debt Auctions** In extreme cases where liquidations fail to raise enough to cover the seized debt, another mechanism kicks in: Debt Auctions. System governance tokens are minted and sold through auction to raise enough stable asset to cover the remaining debt. The governors of the system represent the lenders of last resort.

The system monitors the state of CDPs and debt and triggers these auctions as needed.

## Internal Debt Tracking

Users incur debt when they draw new stable assets from their CDP. Within the system this debt is tracked in the form of a "debt coin" stored internally in the module's accounts. Every time a stable coin is created a corresponding debt coin is created. Likewise when debt is repaid stable coin and internal debt coin are burned.

The cdp module uses two module accounts - one to hold debt coins associated with active CDPs, and another (the "liquidator" account) to hold debt from CDPS that have been seized by the system.

## Fees

When a user repays stable asset withdrawn from a CDP, they must also pay a fee.

This is calculated according to the amount of stable asset withdrawn and the time withdrawn for. Like interest on a loan, fees grow at a compounding percentage of original debt.

Fees create incentives to open or close CDPs and can be changed by governance to help keep the system functioning through changing market conditions.

A further fee is applied on liquidation of a CDP. Normally when the collateral is sold to cover the debt, any excess not sold is returned to the CDP holder. The liquidation fee reduces the amount of excess collateral returned, representing a cut that the system takes.

Fees accumulate to the system and are split between the savings rate and surplus. Fees accumulated by the savings rate are distributed directly to holders of stable coins at a specified frequency. Savings rate distributions are proportional to tokens held. For example, if an account holds 1% of all stable coins, they will receive 1% of the savings rate distribution. Fees accumulated as surplus are automatically sold at auction for governance token once a certain threshold is reached. The governance tokens raised at auction are then burned, acting as incentive for safe governance of the system.

## Governance

The cdp module's behavior is controlled through several parameters which are updated through a governance mechanism. These parameters are listed in [Parameters](04_params.md).

Governance is important for actions such as:

- enabling CDPs to be created with new collateral assets
- changing fee rates to incentivize behavior
- increasing the debt ceiling to allow more stable asset to be created
- increasing/decreasing the savings rate to promote stability of the debt asset

## Dependency: supply

The CDP module relies on a supply keeper to move assets between its module accounts and user accounts.

## Dependency: pricefeed

The CDP module needs to know the current price of collateral assets in order to determine if CDPs are under collateralized. This is provided by a "pricefeed" module that returns a price for a given collateral in units (usually US Dollars) which are the target for the stable asset. The status of the pricefeed for each collateral is checked at the beginning of each block. In the event that the pricefeed does not return a price for a collateral asset:

1. Liquidation of CDPs is suspended until a price is reported
2. Accumulation of fees is suspended until a price is reported
3. Deposits and withdrawals of collateral are suspended until a price is reported
4. Creation of new CDPs is suspended until a price is reported
5. Drawing of additional debt off of existing CDPs is suspended until a price is reported
