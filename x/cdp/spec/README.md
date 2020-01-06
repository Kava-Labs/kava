
# `cdp`

## Table of Contents

<!-- TOC -->
1. **[Concepts](01_concepts.md)**
2. **[State](02_state.md)**
3. **[State Transistions](03_state_transistsions.md)**
4. **[Messages](04_messages.md)**
5. **[EndBlock](05_end_block.md)**
6. **[Events](06_events.md)**
7. **[Parameters](07_params.md)**
8. **[EndBlock](08_future_improvements.md)**



## Overview

The `x/cdp` module stores and manages Collateralized Debt Positions (or CDPs). Through this module CDPs can be created, modified, and removed by users. This module does not handle system stability through liquidation of CDPs and creation of auctions; relying on other modules to fulfill this role. Further this module requires a pricefeed to determine if CDPs can be updated or not.

User interactions with this module:

- create a new cdp by depositing some type of coin as collateral
- withdraw newly minted stable coin from this CDP (up to a fraction of the value of the collateral)
- repay debt by paying back stable coins (including paying any fees accrued)
- remove collateral and close CDP

Notable features

- CDPs are multi collateral, but limited to one collateral per cdp.
- CDPs are owned by one or more addresses, each with the authorization to add and remove collateral and stable coin.
- CDPs support multiple stable coin denominations.
