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
6. **[Hooks](06_hooks.md)**
7. **[BeginBlock](07_begin_block.md)**

## Abstract

`x/incentive` is an implementation of a Cosmos SDK Module that allows for governance controlled user incentives for users who take certain actions, such as opening a collateralized debt position (CDP). Governance proposes an array of rewards, with each item representing a collateral type that will be eligible for rewards. Each collateral reward specifies the number of coins awarded per second, the length of rewards periods. Governance can alter the collateral rewards using parameter change proposals as well as adding or removing collateral types. All changes to parameters would take place in the _next_ period. User rewards are __opt in__, ie. users must claim rewards in order to receive them. If users fail to claim rewards before the claim period expiry, they are no longer eligible for rewards.

### Dependencies

This module uses hooks to update user rewards. Currently, `incentive` implements hooks from the `cdp`, `hard`, `swap`, and `staking` (comsos-sdk) modules. All rewards are paid out from the `kavadist` module account.
