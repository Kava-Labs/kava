
# `committee`

## Table of Contents

## Overview

The `x/committee` module is an additional governance module to `cosmos-sdk/x/gov`.

It allows groups of accounts to vote on and enact proposals, mainly to allow certain proposal types to be decided on quickly in emergency situations, or to delegate low risk parameter updates to a smaller group of individuals.

Committees have members and permissions.

Members vote on proposals, with just simple one vote per member, no deposits or slashing. More sophisticated voting could be added.

A permission acts as a filter for incoming gov proposals, rejecting them if they do not pass. A permission can be anything with a method `Allows(p Proposal) bool`. They reject all proposals that they don't explicitly allow.

This allows permissions to be parameterized to allow fine grained control specified at runtime. For example a generic parameter permission type can exist, but then on a live chain a permission can be added to a group to allow them to only change a particular param, even restricting the range of allowed change.

Design Alternatives

- Should this define its own gov types, or reuse those from gov module?
- Should we push changes to sdk gov to make it more general purpose?
- Could use params more instead of custom gov proposals
