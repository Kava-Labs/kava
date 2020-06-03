<!--
order: 0
title: "Committee Overview"
parent:
  title: "committee"
-->

# `committee`

## Table of Contents

<!-- TOC -->
1. **[Concepts](01_concepts.md)**
2. **[State](02_state.md)**
3. **[Messages](03_messages.md)**
4. **[Events](04_events.md)**
5. **[Params](05_params.md)**
6. **[BeginBlock](06_begin_block.md)**

## Overview

The `x/committee` module is Cosmos-SDK module that acts as an additional governance module to `cosmos-sdk/x/gov`.

It allows groups of accounts to vote on and enact proposals without a full chain governance vote. Certain proposal types can then be decided on quickly in emergency situations, or low risk parameter updates can be delegated to a smaller group of individuals.

Committees work with "proposals", using the same type from the `gov` module so they are compatible with all existing proposal types such as param changes, or community pool spend, or text proposals.

Committees have members and permissions. Committees are 'elected' via traditional `gov` proposals - ie. all coin-holders vote on the creation, deletion, and updating of committees.

Members of committees vote on proposals, with one vote per member and no deposits or slashing. Only a member of a committee can submit a proposal for that committee. More sophisticated voting could be added, as well as the ability for committees to edit themselves or other committees. A proposal passes when the number of votes is over the threshold for that committee. Vote thresholds are set per committee. Committee members vote yes by casting a vote and vote no by abstaining from voting and letting the proposal expire.

Permissions scope the allowed set of proposals a committee can enact. For example:

- allow the committee to only change the cdp `CircuitBreaker` param.
- allow the committee to change auction bid increments, but only within the range [0, 0.1]
- allow the committee to only disable cdp msg types, but not staking or gov

A permission acts as a filter for incoming gov proposals, rejecting them at the handler if they do not have the required permissions. A permission can be any type with a method `Allows(p Proposal) bool`. The handler will reject all proposals that are not explicitly allowed. This allows permissions to be parameterized to allow fine grained control specified at runtime. For example a generic parameter permission type can allow a committee to only change a particular param, or only change params within a certain range.
