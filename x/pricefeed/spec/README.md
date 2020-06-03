<!--
order: 0
title: "Pricefeed Overview"
parent:
  title: "pricefeed"
-->

# `pricefeed`

<!-- TOC -->
1. **[Concepts](01_concepts.md)**
2. **[State](02_state.md)**
3. **[Messages](03_messages.md)**
4. **[Events](04_events.md)**
5. **[Params](05_params.md)**
6. **[EndBlock](06_end_block.md)**

## Abstract

`x/pricefeed` is an implementation of a Cosmos SDK Module that handles the posting of prices for various markets by a group of whitelisted oracles. At the end of each block, the median price of all oracle posted prices is determined for each market and stored.
