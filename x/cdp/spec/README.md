
# `cdp`

## Table of Contents

1. **[Concepts](01_concepts.md)**
2. **[State](02_state.md)**
3. **[Messages](03_messages.md)**
4. **[BeginBlock](04_begin_block.md)**
5. **[Events](05_events.md)**
6. **[Parameters](06_params.md)**

## Overview

The `x/cdp` module stores and manages Collateralized Debt Positions (or CDPs).

A CDP enables the creation of an asset pegged to an external price (usually US Dollar) by collateralization with another asset. Collateral is locked in a CDP and new pegged asset can be minted up to approximately the value of the collateral. To unlock the collateral, the debt must be repaid by returning some pegged asset to the CDP at which point it will be burned and the collateral unlocked.

Pegged assets remain fully collateralized by the value locked in CDPs. In the event of price changes some of this collateral can be seized and sold off by the system to reclaim and reduce the supply of pegged assets.
