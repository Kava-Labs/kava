<!--
order: 0
title: "Auction Overview"
parent:
  title: "auction"
-->

# `auction`

<!-- TOC -->
1. **[Concepts](01_concepts.md)**
2. **[State](02_state.md)**
3. **[Messages](03_messages.md)**
4. **[Events](04_events.md)**
5. **[Params](05_params.md)**
6. **[BeginBlock](06_begin_block.md)**

## Abstract

`x/auction` is an implementation of a Cosmos SDK Module that handles the creation, bidding, and payout of 3 distinct auction types. All auction types implement the `Auction` interface. Each auction type is used at different points during the normal functioning of the CDP system.
