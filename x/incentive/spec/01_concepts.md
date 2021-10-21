<!--
order: 1
-->

# Concepts

This module presents an implementation of user incentives that are controlled by governance. When users take a certain action, for example opening a CDP, they become eligible for rewards. Rewards are **opt in** meaning that users must submit a message before the claim deadline to claim their rewards. The goals and background of this module were subject of a previous Kava governance proposal, which can be found [here](https://ipfs.io/ipfs/QmSYedssC3nyQacDJmNcREtgmTPyaMx2JX7RNkMdAVkdkr/user-growth-fund-proposal.pdf)

## General Reward Distribution

Rewards target various user activity. For example, usdx borrowed from bnb CDPs, btcb supplied to the hard money market, or owned shares in a swap kava/usdx pool.

Each second, the rewards accumulate at a rate set in the params, eg 100 ukava per second. These are then distributed to all users ratably based on their percentage involvement in the rewarded activity. For example if a user holds 1% of all funds deposited to the kava/usdx swap pool. They will receive 1% of the total rewards each second.

The quantity tracking a user's involvement is referred to as "source shares" in the code. And the total across all users the "total source shares". The quotient then gives their percentage involvement, eg if a user borrowed 10,000 usdx, and there is 100,000 usdx borrowed by all users, then they will get 10% of rewards.

## Efficiency

Paying out rewards to every user every block would be slow and lead to long block times. Instead rewards are calculated much less frequently.

Every block a global tracker adds up total rewards paid out per unit of user involvement. For example, per unit of xrpb supplied to hard, or per share in a kava/usdx swap pool. A user's specific reward can then be calculated as needed based on their current source shares.

Users' rewards must be updated whenever their source shares change. This happens through hooks into other modules that run before deposits/borrows/supplies etc.

## HARD Token distribution

The incentive module also distributes the HARD token on the Kava blockchain. HARD tokens are distributed to two types of ecosystem participants:

1. Kava stakers - any address that stakes (delegates) KAVA tokens will be eligible to claim HARD tokens. For each delegator, HARD tokens are accumulated ratably based on the total number of kava tokens staked. For example, if a user stakes 1 million KAVA tokens and there are 100 million staked KAVA, that user will accumulate 1% of HARD tokens earmarked for stakers during the distribution period. Distribution periods are defined by a start date, an end date, and a number of HARD tokens that are distributed per second.
2. Depositors/Borrows - any address that deposits and/or borrows eligible tokens to the hard module will be eligible to claim HARD tokens. For each depositor, HARD tokens are accumulated ratably based on the total number of tokens staked of that denomination. For example, if a user deposits 1 million "xyz" tokens and there are 100 million xyz deposited, that user will accumulate 1% of HARD tokens earmarked for depositors of that denomination during the distribution period. Distribution periods are defined by a start date, an end date, and a number of HARD tokens that are distributed per second.

Users are not air-dropped tokens, rather they accumulate `Claim` objects that they may submit a transaction in order to claim. In order to better align long term incentives, when users claim HARD tokens, they have options, called 'multipliers', for how tokens are distributed.

The exact multipliers will be voted by governance and can be changed via a governance vote. An example multiplier schedule would be:

- Short-term locked - 20% multiplier and 1 month transfer restriction. Users receive 20% as many tokens as users who choose long-term locked tokens.
- Long-term locked - 100% multiplier and 1 year transfer restriction. Users receive 5x as many tokens as users who choose short-term locked tokens.

## USDX Minting Rewards

The incentive module is responsible for distribution of KAVA tokens to users who mint USDX. When governance adds a collateral type to be eligible for rewards, they set the rate (coins/second) at which rewards are given to users, the length of each reward period, the length of each claim period, and the amount of time reward coins must vest before users who claim them can transfer them. For the duration of a reward period, any user that has minted USDX using an eligible collateral type will ratably accumulate rewards in a `USDXMintingClaim` object. For example, if a user has minted 10% of all USDX for the duration of the reward period, they will earn 10% of all rewards for that period. When the reward period ends, the claim period begins immediately, at which point users can submit a message to claim their rewards. Rewards are time-locked, meaning that when a user claims rewards they will receive them as a vesting balance on their account. Vesting balances can be used to stake coins, but cannot be transferred until the vesting period ends. In addition to vesting, rewards can have multipliers that vary the number of tokens received. For example, a reward with a vesting period of 1 month may have a multiplier of 0.25, meaning that the user will receive 25% of the reward balance if they choose that vesting schedule.

## SWP Token Distribution

The incentive module distributes the SWP token on the Kava blockchain. SWP tokens are distributed to two types of ecosystem participants:

1. Kava stakers - any address that stakes (delegates) KAVA tokens will be eligible to claim SWP tokens. For each delegator, SWP tokens are accumulated ratably based on the total number of kava tokens staked. For example, if a user stakes 1 million KAVA tokens and there are 100 million staked KAVA, that user will accumulate 1% of SWP tokens earmarked for stakers during the distribution period. Distribution periods are defined by a start date, an end date, and a number of SWP tokens that are distributed per second.
2. Liquidity providers - any address that provides liquidity to eligible Swap protocol pools will be eligible to claim SWP tokens. For each liquidity provider, SWP tokens are accumulated ratably based on the total amount of pool shares. For example, if a liquidity provider deposits "xyz" and "abc" tokens into the "abc:xyz" pool to receive 10 shares and the pool has 50 total shares, then that user will accumulate 20% of SWP tokens earmarked for liquidity providers of that pool during the distribution period. Distribution periods are defined by a start date, an end date, and a number of SWP tokens that are distributed per second.
