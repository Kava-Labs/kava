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

## Overview

The `evmutil` module provides additional functionalities on top of the `evm` module by storing additional state data for evm accounts and exposing an `EvmBankKeeper` that should be used by the `x/evm` keeper for bank operations.

Two keepers are exposed by the module, `EvmBankKeeper` and the module `Keeper`.

The main purpose of the `EvmBankKeeper` is to allow the usage of the `akava` balance on the EVM through an account's existing `ukava` balance.
This is needed because the EVM gas token use 18 decimals, and since `ukava` has 6 decimals, it cannot be used as the EVM gas denom directly.

The module `Keeper` provides access to an account's excess `akava` balance and the ability to update `akava` balances. This is needed by the `EvmBankKeeper` to store
excess `akava` balances since the `ukava` token only has 6 decimals of precision.
