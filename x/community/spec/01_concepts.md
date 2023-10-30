<!--
order: 1
-->

# Concepts

## Community Pool

The x/community module contains the community pool funds and provides proposal
handlers to manage community pool funds.


The x/community module facilitates interactions with the community pool. In a future release, the community pool may be fully replaced by the x/community module account, but for now it is a place to keep things like proposals for interacting with the vanilla cosmos-sdk community pool (a portion of the auth fee pool held by the x/distribution module account).

### Funding

The x/community module account can be funded by any account sending a
community/FundCommunityPool message. These funds may be deposited/withdrawn to
lend via the CommunityPoolLendDepositProposal &
CommunityPoolLendWithdrawProposal.

### Rewards

Staking rewards are paid out every block from the community pool, instead of
from minted coins from `x/mint` and `x/kavadist`. The payout is calculated with
the set `StakingRewardsPerSecond` in the module parameters.
