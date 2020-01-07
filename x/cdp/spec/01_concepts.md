# Concepts

## Collateralized Debt Positions

CDPs enable the creation of a stable pegged asset by collateralization with another on chain asset.

A CDP is owned by a set of users (addresses) and stores one type of collateral. Any user can deposit, withdraw, collateral and pegged asset from their CDPs.

User interactions with this module:

- create a new cdp by depositing some type of coin as collateral
- withdraw newly minted stable coin from this CDP (up to a fraction of the value of the collateral)
- repay debt by paying back stable coins (including paying any fees accrued)
- remove collateral and close CDP

This module does not handle system stability through liquidation of CDPs and sell off of collateral; instead relying on other modules to fulfill this role. Further an up to date price of each collateral is requited to determine their values.

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

## Dependency: price feed

The CDP module needs to know the current real world price of collateral assets in order to calculate if CDPs are under collateralized. It calls another module to return a price for a given collateral in units of the pegged asset (usually US Dollars).

CDP assumes the price feed is accurate, leaving it to external modules to achieve this.

## Dependency: liquidator

The CDP module on its own cannot maintain system stability. In the event the system becomes under collateralized, a re-balancing mechanism is needed to seize users' collateral and sell it for the stable asset. Additional stability mechanisms such as minting or burning governance tokens can also be used.

This module exposes the concept of CDP seizing through a keeper method, but does not handle the selling of the seized assets. CDPs can be seized when their collateralization ratio falls below a threshold (usually > 100%) set in the system parameters.
