<!--
order: 1
-->

# Concepts

This module presents an implementation of user incentives that are controlled by governance. When users take a certain action, in this case opening a CDP, they become eligible for rewards. Rewards are __opt in__ meaning that users must submit a message before the claim deadline to claim their rewards. The goals and background of this module were subject of a previous Kava governance proposal, which can be found [here](https://ipfs.io/ipfs/QmSYedssC3nyQacDJmNcREtgmTPyaMx2JX7RNkMdAVkdkr/user-growth-fund-proposal.pdf).

When governance adds a collateral type to be eligible for rewards, they set the rate (coins/time) at which rewards are given to users, the length of each reward period, the length of each claim period, and the amount of time reward coins must vest before users who claim them can transfer them. For the duration of a reward period, any user that has minted USDX using an eligible collateral type will ratably accumulate rewards in a `Claim` object. For example, if a user has minted 10% of all USDX for the duration of the reward period, they will earn 10% of all rewards for that period. When the reward period ends, the claim period begins immediately, at which point users can submit a message to claim their rewards. Rewards are time-locked, meaning that when a user claims rewards they will receive them as a vesting balance on their account. Vesting balances can be used to stake coins, but cannot be transferred until the vesting period ends.
