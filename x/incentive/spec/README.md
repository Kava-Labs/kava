<!--
order: 0
title: "Incentive Overview"
parent:
  title: "incentive"
-->

# `incentive`

<!-- TOC -->
1. **[Concepts](01_concepts.md)**
2. **[State](02_state.md)**
3. **[Messages](03_messages.md)**
4. **[Events](04_events.md)**
5. **[Params](05_params.md)**
6. **[BeginBlock](06_begin_block.md)**

## Abstract

`x/incentive` is an implementation of a Cosmos SDK Module that allows for governance controlled user incentives for users who create stable assets by opening a collateralized debt position (CDP). Governance proposes an array of collateral rewards, with each item representing a collateral type that will be eligible for rewards. Each collateral reward specifies the number of coins awarded per period, the length of rewards periods, the length of claim periods. Governance can alter the collateral rewards using parameter change proposals as well as adding or removing collateral types. All changes to parameters would take place in the _next_ period.

### Dependencies

This module depends on `x/cdp` for users to be able to create CDPs and on `x/kavadist`, which controls the module account from where rewards are spent. In the event that the module account is not funded, user's attempt to claim rewards will fail.
