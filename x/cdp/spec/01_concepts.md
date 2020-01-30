# Concepts

## Collateralized Debt Positions

CDPs enable the creation of a stable asset by collateralization with another on chain asset.

A CDP is scoped to one collateral type. It has one primary owner, and a set of "depositors". The depositors can deposit and withdraw collateral to the CDP. The owner can draw stable assets (creating debt) and repay them to cancel the debt.

Once created stable assets are free to be transferred between users, but a CDP owner must repay their debt to get their collateral back.

User interactions with this module:

- create a new cdp by depositing some type of coin as collateral
- withdraw newly minted stable coin from this CDP (up to a fraction of the value of the collateral)
- repay debt by paying back stable coins (including paying any fees accrued)
- remove collateral and close CDP

## Liquidation & Stability System

In the event of a decrease in the price of the collateral, the total value of all collateral in CDPs may drop below the value of all the issued stable assets. This undesirable event is countered through two mechanisms:

**CDP Liquidations** The ratio of collateral value to debt value in each CDP is monitored. When this drops too low the collateral and debt is automatically seized by the system. The collateral is sold off through an auction to bring in stable asset which is burned against the seized debt.

**Debt Auctions** In extreme cases where liquidations fail to raise enough to cover the seized debt, another mechanism kicks in: Debt Auctions. System governance tokens are minted and sold through auction to raise enough stable asset to cover the remaining debt. The governors of the system represent the lenders of last resort.

The system monitors the state of CDPs and debt and triggers these auctions as needed.

## Internal Debt Tracking

Users incur debt when they draw new stable assets from their CDP. Within the system this debt is tracked in the form of a "debt coin" stored internally in the module's accounts. Every time a stable coin is created a corresponding debt coin is created. Likewise when debt is repaid stable coin and internal debt coin are burned.

The cdp module uses two module accounts - one to hold debt coins associated with active CDPs, and another (the "liquidator" account) to hold debt from CDPS that have been seized by the system.

## Fees

When a user repays stable asset withdrawn from a CDP, they must also pay a fee.

This is calculated according to the amount of stable asset withdrawn and the time withdrawn for. Like interest on a loan fees grow at a compounding percentage of original debt.

Fees create incentives to open or close CDPs and can be changed by governance to help keep the system functioning through changing market conditions.

A further fee is applied on liquidation of a CDP. Normally when the collateral is sold to cover the debt, any excess not sold is returned to the CDP holder. The liquidation fee reduces the amount of excess collateral returned, representing a cut that the system takes.

Fees accumulate to the system before being automatically sold at auction for governance token. These are then burned, acting as incentive for safe governance of the system.

## Governance

The cdp module's behavior is controlled through several parameters which are updated through a governance mechanism. These parameters are listed in [Parameters](06_params.md).

Governance is important for actions such as:

- enabling CDPs to be created with new collateral assets
- changing fee rates to incentivize behavior
- increasing the debt ceiling to allow more stable asset to be created

## Dependency: supply

The CDP module relies on a supply keeper to move assets between its module accounts and user accounts.

## Dependency: pricefeed

The CDP module needs to know the current price of collateral assets in order to determine if CDPs are under collateralized. This is provided by a "pricefeed" module that returns a price for a given collateral in units (usually US Dollars) which are the target for the stable asset.
