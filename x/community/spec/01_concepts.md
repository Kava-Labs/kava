<!--
order: 1
-->

# Concepts

## Community Pool

The x/community module facilitates interactions with the community pool. In a future release, the community pool may be fully replaced by the x/community module account, but for now it is a place to keep things like proposals for interacting with the vanilla cosmos-sdk community pool (a portion of the auth fee pool held by the x/distribution module account).

### Funding

The x/community module account can be funded by any account sending a community/FundCommunityPool message. These funds are not currently used for anything. The module account is also the transitory passthrough account for community pool funds that are deposited/withdrawn to lend via the CommunityPoolLendDepositProposal & CommunityPoolLendWithdrawProposal.
