# `x/precisebank`

## Abstract

This document specifies the precisebank module of Kava.

The precisebank module is responsible for extending the precision of `x/bank`,
intended to be used for the `x/evm`. It serves as a wrapper of `x/bank` to
increase the precision of KAVA from 6 to 18 decimals, while preserving the
behavior of existing `x/bank` balances.

This module is used only by `x/evm` where 18 decimal points are expected.
