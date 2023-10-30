<!--
order: 1
-->

# Concepts

## Community Pool

The x/community module contains the community pool funds and provides proposal
handlers to manage community pool funds.

### Funding

The x/community module account can be funded by any account sending a
community/FundCommunityPool message. These funds may be deposited/withdrawn to
lend via the CommunityPoolLendDepositProposal &
CommunityPoolLendWithdrawProposal.

### Rewards

Rewards payout behavior for staking depends on the module parameters, and will
change based on the "switchover" time parameter `upgrade_time_disable_inflation`.

If the current block is *before* the switchover time and the
`staking_rewards_per_second` parameter is set to 0, no staking rewards will be
paid from the `x/community` module and will continue to come from other modules
such as `x/mint` and `x/distribution`.

On the first block after the switchover time, the `staking_rewards_per_second`
parameter is updated to reflect the parameter
`upgrade_time_set_staking_rewards_per_second`, and staking rewards are paid out
every block from the community pool, instead of from minted coins from `x/mint`
and `x/kavadist`. The payout is calculated with the `staking_rewards_per_second`
parameter.

In addition to these payout changes, inflation in `x/mint` and `x/kavadist` is
disabled after the switchover time.
