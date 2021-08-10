<!--
order: 1
-->

# Concepts

For a general introduction to governance using the Comsos-SDK, see [x/gov](https://github.com/cosmos/cosmos-sdk/blob/v0.38.3/x/gov/spec/01_concepts.md).

This module provides companion governance functionality to `x/gov` by allowing the creation of committees, or groups of addresses that can vote on proposals for which they have permission and which bypass the usual on-chain governance structures. Permissions scope the types of proposals that committees can submit and vote on. This allows for committees with unlimited breadth (ie, a committee can have permission to perform any governance action), or narrowly scoped abilities (ie, a committee can only change a single parameter of a single module within a specified range).

Committees are either member committees governed by a set of whitelisted addresses or token committees whose votes are weighted by token balance. For example, the [Kava Stability Committee](https://medium.com/kava-labs/kava-improves-governance-enabling-faster-response-to-volatile-markets-2d0fff6e5fa9) is a member committee that has the ability to protect critical protocol infrastructure by briefly pausing certain functionality; while the Hard Token Committee allows HARD token holders to participate in governance related to HARD protocol on the Kava blockchain. Further, committees can tally votes by either the "first-past-the-post" or "deadline" tallying procedure. Committees with "first-past-the-post" vote tallying enact proposals immediately once they pass, allowing greater flexibility than permitted by `x/gov`. Committees with "deadline" vote tallying evaluate proposals at their deadline, allowing time for all stakeholders to vote before a proposal is enacted or rejected.
