<!--
order: 0
title: Evmutil Overview
parent:
  title: "evmutil"
-->

# `evmutil`

## Table of Contents

<!-- TOC -->

1. **[Concepts](01_concepts.md)**
2. **[State](02_state.md)**
3. **[Messages](03_messages.md)**
4. **[Events](04_events.md)**
5. **[Params](05_params.md)**

## Overview

The evmutil module provides additional functionalities on top of the evm module.

### EVM `akava` Usage

evmutil stores additional state data for evm accounts and exposes an `EvmBankKeeper` that should be used by the `x/evm` keeper for bank operations.
The purpose of the `EvmBankKeeper` is to allow the usage of the `akava` balance on the EVM via an account's existing `ukava` balance. This is needed because the EVM gas token use 18 decimals, and since `ukava` has 6 decimals, it cannot be used as the EVM gas denom directly.

For additional details on how balance conversions work, see **[Concepts](01_concepts.md)**.

### ERC20 Token <> sdk.Coin Conversion

evmutil exposes messages to allow for the conversion of Kava ERC20 tokens and sdk.Coins via a whitelist.

For additional details on how these messages work, see **[Messages](03_messages.md)**.
