<!--
order: 1
-->

# Concepts

This module implements governance controlled user incentives. When users take a certain action, for example opening a CDP, they become eligible for rewards. Rewards are **opt in** meaning that users must submit a message before the claim deadline to claim their rewards. The goals and background of this module were subject of a previous Kava governance proposal, which can be found [here](https://ipfs.io/ipfs/QmSYedssC3nyQacDJmNcREtgmTPyaMx2JX7RNkMdAVkdkr/user-growth-fund-proposal.pdf)

## General Reward Distribution

Rewards target various user activity. For example, usdx borrowed from bnb CDPs, btcb supplied to the hard money market, or shares owned in a swap kava/usdx pool.

Each second, the rewards accumulate at a rate set in the params, eg 100 ukava per second. These are then distributed to all users ratably based on their percentage involvement in the rewarded activity. For example if a user holds 1% of all funds deposited to the kava/usdx swap pool. They will receive 1% of the total rewards each second.

The quantity tracking a user's involvement is referred to as "source shares". And the total across all users the "total source shares". The quotient then gives their percentage involvement, eg if a user borrowed 10,000 usdx, and there is 100,000 usdx borrowed by all users, then they will get 10% of rewards.

## Efficiency

Paying out rewards to every user every block would be slow and lead to long block times. Instead rewards are calculated lazily only when needed.

First, every block, the amount of rewards to be distributed in that block are divided by the total source shares to get the rewards per share. This is added to a global total (named "global indexes"). This is repeated every block such that the global indexes represents the total rewards a user should be owed per source share if they had held a deposit from when the rewards were created.

Then, if a user has deposited (say into a CDP) at the very start of the chain (and never changed their deposit), their current reward balance can be calculated at any time $t$ as

$$
\texttt{rewards}_ t = \texttt{globalIndexes}_ t \cdot \texttt{sourceShares}_ t
$$

If a user modifies their source shares (at say time $t-10$) we can still calculate their total rewards:

$$
\texttt{rewards}_ t= \text{rewards accrued up to time t-10} + \text{rewards accrued from time t-10 to time t}
$$

$$
\texttt{rewards}_ t = \texttt{globalIndexes}_ {t-10} \cdot \texttt{sourceShares}_ {t-10} + (\texttt{globalIndexes}_ t - \texttt{globalIndexes}_ {t-10}) \cdot \texttt{sourceShares}_ t
$$

This generalizes to any number of source share modifications.

In code, to avoid storing the entire history of a user's source shares and global index values, rewards are calculated on every source shares change and added to a reward balance:

$$
\texttt{rewards}_ t = \texttt{rewardBalance}_ {t -10} + (\texttt{globalIndexes}_ t - \texttt{globalIndexes}_ {t-10}) \cdot \texttt{sourceShares}_ t
$$

Old values of $\texttt{rewardBalance}$ and $\texttt{globalIndexes}$ ares stored in a `Claim` object for each user as `rewardBalance` and `rewardIndexes` respectively.

Listeners on external modules fire to update these values when source shares change. For example, when a user deposits to hard, a method in incentive is called. This fundamental operation is called "sync". It calculates the rewards accrued since last time the `sourceShares` changed, adds it to the claim, and stores the current `globalIndexes` in the `rewardIndexes`. Sync must be called whenever source shares change, otherwise incorrect rewards will be distributed.

Enumeration of 'sync' input states:
- `sourceShares`, `globalIndexes`, or `rewardIndexes` should never be negative
- `globalIndexes` >= `rewardIndexes` (global indexes must never decrease)
- `globalIndexes` and `rewardIndexes` can be positive or 0, where not existing in the store is counted as 0

- `sourceShares` are the value before the update (eg before a hard deposit)

 | `globalIndexes` | `rewardIndexes` | `sourceShares` | description                                                                                                                                                                                                                                                |
 |------------------|-----------------|-----------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
 | positive         | positive        | positive        | normal sync                                                                                                                                                                                                                                                |
 | positive         | positive        | 0               | normal (can happen when a user is creating a deposit (so shares are increasing from 0))                                                                                                                                                                    |
 | positive         | 0               | positive        | The claim doesn't hold indexes, so the global indexes must have been added since last sync (eg a new denom was added to reward params). This is indistinguishable from a claim accidentally being deleted, where it will accrue a large amount of rewards. |
 | positive         | 0               | 0               | User is creating source shares.                                                                                                                                                                                                                            |
 | 0                | positive        | positive        | global indexes < claim indexes - fatal error, otherwise the new rewards will be negative                                                                                                                                                                   |
 | 0                | positive        | 0               | global indexes < claim indexes - fatal error                                                                                                                                                                                                               |
 | 0                | 0               | positive        | Source has no rewards yet. User is updating their shares.                                                                                                                                                                                                  |
 | 0                | 0               | 0               | Source has no rewards yet. User is creating source shares.                                                                                                                                                                                                 |

It is important that:
- claim indexes are not deleted
    - Otherwise when sync is called, it will fill them in with 0 values and perform sync as if the user had deposit since the beginning of the rewards (usually accumulating a lot of rewards).
- global indexes are not deleted
    - Otherwise claims cannot be synced. Problematic if a sync happens in a begin blocker and it panics.
- hooks are called any time source shares change
    - If source shares can be updated without a sync, it can be possible to accumulate far too much rewards. For example, a user who holds a small deposit for a long time could deposit a large amount and skip the sync, then trigger a sync which will calculate rewards as if the large deposit was there for a long time.

The code is further complicated by:
- Claim objects contain indexes for several source shares.
- Rewards for hard borrows and hard deposits use the same claim object.
- Savings and hard hooks trigger any time one in a group of source shares change, but don't identify which changed.
- The hard `BeforeXModified` hooks don't show source shares that have increased from zero (eg when a new denom is deposited to an existing deposit). So there is an additional `AfterXModified` hook, and the claim indexes double up as a copy of the borrow/deposit denoms.
- The sync operation is split between two methods to try to protect against indexes being deleted.
    - `InitXRewards` performs a sync assuming source shares are 0, it mostly fires in cases where `sourceShares` = 0 above (except for hard and supply)
    - `SyncXRewards` performs a sync, but skips it if `globalIndexes` are not found or `rewardIndexes` are not found (only when claim object not found)
- Usdx rewards do not support multiple reward denoms.

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
