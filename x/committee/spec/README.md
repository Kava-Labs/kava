
# `committee`

## Table of Contents

## Overview

The `x/committee` module is an additional governance module to `cosmos-sdk/x/gov`.

It allows groups of accounts to vote on and enact proposals without a full chain governance vote. Certain proposal types can then be decided on quickly in emergency situations, or low risk parameter updates can be delegated to a smaller group of individuals.

Committees work with "proposals", using the same type from the `gov` module so they are compatible with all existing proposal types such as param changes, or community pool spend, or text proposals.

Committees have members and permissions.

Members vote on proposals, with just simple one vote per member, no deposits or slashing. More sophisticated voting could be added.

Permissions scope the allowed set of proposals a committee can enact. For example:

- allow the committee to only change the cdp `CircuitBreaker` param.
- allow the committee to change auction bid increments, but only within the range [0, 0.1]
- allow the committee to only disable cdp msg types, but not staking or gov

A permission acts as a filter for incoming gov proposals, rejecting them if they do not pass. A permission can be any type with a method `Allows(p Proposal) bool`. They reject all proposals that they don't explicitly allow.

This allows permissions to be parameterized to allow fine grained control specified at runtime. For example a generic parameter permission type can allow a committee to only change a particular param, or only change params within a certain percentage.
