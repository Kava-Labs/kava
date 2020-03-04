
# `shutdown`

## Table of Contents

## Overview

The `x/shutdown` module allows certain message types to be disabled based on governance votes.

Msgs and routes are disabled via an antehandler decorator. The decorator checks incoming all txs and rejects them if they contain a disallowed msg type.
Disallowed msg types are stored in a circuit breaker keeper.

The list of disallowed msg types is updated via a custom governance proposal and handler.

Design Alternatives:

- store list of disallowed msg types in params, then we don't need the custom gov proposal
- replace the app Router with a custom one to avoid using the antehandler - can't be done with current baseapp, but v0.38.x enables this. (https://github.com/cosmos/cosmos-sdk/issues/5455)