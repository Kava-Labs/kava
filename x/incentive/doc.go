/*
Package incentive implements a Cosmos SDK module, per ADR 009, that provides governance-controlled,
on-chain incentives for users who open cdps and mint stablecoins (USDX).

For the background and motivation of this module, see the governance proposal that was voted on by
KAVA token holders: https://ipfs.io/ipfs/QmSYedssC3nyQacDJmNcREtgmTPyaMx2JX7RNkMdAVkdkr/user-growth-fund-proposal.pdf

The 'Reward' parameter is used to control how much incentives are given.
For example, the following reward:

	Reward{
		Active: true,
		Denom: "bnb",
		AvailableRewards: sdk.NewCoin("ukava", 1000000000),
		Duration: time.Hour*7*24,
		TimeLock: time.Hour*24*365,
		ClaimDuration: time.Hour*7*24,
	}

will distribute 1000 KAVA each week (Duration) to users who mint USDX using collateral bnb.
That KAVA can be claimed by the user for one week (ClaimDuration) after the reward period expires,
and all KAVA rewards will be timelocked for 1 year (TimeLock). If  a user does not claim them during
the claim duration period, they are forgone.

Rewards are accumulated by users continuously and proportionally - ie. if a user holds a CDP that has
minted 10% of all USDX backed by bnb for the entire reward period, they will be eligible to claim 10%
of rewards for that period.

Once a reward period ends, but not before, users can claim the rewards they have accumulated.  Users claim rewards
using a MsgClaimReward transaction. The following msg:

	MsgClaimReward {
		"kava1..."
		"bnb"
	}

will claim all outstanding rewards for minting USDX backed by bnb for the input user.

*/
package incentive
