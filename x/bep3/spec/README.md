# `bep3` module specification

<!-- TOC -->
1. **[Concepts](01_concepts.md)**
2. **[State](02_state.md)**
3. **[Messages](03_messages.md)**
4. **[Events](04_events.md)**
5. **[Params](05_params.md)**
6. **[BeginBlock](06_begin_block.md)**

## Abstract

`x/bep3` is an implementation of a Cosmos SDK Module that handles cross-chain Atomic Swaps between Kava and blockchains that implement the BEP3 protocol. Atomic Swaps are created, then either claimed before their expiration block or refunded after they've expired.
