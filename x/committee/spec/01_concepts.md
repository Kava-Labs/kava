# Concepts

For a general introduction to governance using the Comsos-SDK, see [x/gov](https://github.com/cosmos/cosmos-sdk/blob/v0.38.3/x/gov/spec/01_concepts.md).

This module provides companion governance functionality to `x/gov` by allowing the creation of committees, or groups of addresses that can vote on proposals for which they have permission and which bypass the usual on-chain governance structures. Permissions scope the types of proposals that committees can submit and vote on. This allows for committees with unlimited breadth (ie, a committee can have permission to perform any governance action), or narrowly scoped abilities (ie, a committee can only change a single parameter of a single module within a specified range). Further, vote tallying is "first-past-the-post", so proposals can be enacted more rapidly and with greater flexibility than permitted by `x/gov`.
