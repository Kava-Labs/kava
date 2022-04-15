package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/x/incentive/types"
	savingstypes "github.com/kava-labs/kava/x/savings/types"
)

// SynchronizeSavingsRewardTests runs unit tests for the keeper.SynchronizeSavingsReward method
type SynchronizeSavingsRewardTests struct {
	unitTester
}

func TestSynchronizeSavingsReward(t *testing.T) {
	suite.Run(t, new(SynchronizeSavingsRewardTests))
}

func (suite *SynchronizeSavingsRewardTests) TestClaimUpdatedWhenGlobalIndexesHaveIncreased() {
	// This is the normal case
	// Given some time has passed (meaning the global indexes have increased)
	// When the claim is synced
	// The user earns rewards for the time passed, and the claim indexes are updated

	originalReward := arbitraryCoins()
	denom := "test"

	claim := types.SavingsClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner:  arbitraryAddress(),
			Reward: originalReward,
		},
		RewardIndexes: types.MultiRewardIndexes{
			{
				CollateralType: denom,
				RewardIndexes: types.RewardIndexes{
					{
						CollateralType: "rewarddenom",
						RewardFactor:   d("1000.001"),
					},
				},
			},
		},
	}
	suite.storeSavingsClaim(claim)

	globalIndexes := types.MultiRewardIndexes{
		{
			CollateralType: denom,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "rewarddenom",
					RewardFactor:   d("2000.002"),
				},
			},
		},
	}
	suite.storeGlobalSavingsIndexes(globalIndexes)

	userShares := i(1e9)
	deposit := savingstypes.NewDeposit(claim.Owner, sdk.NewCoins(sdk.NewCoin(denom, userShares)))
	suite.keeper.SynchronizeSavingsReward(suite.ctx, deposit, getDenoms(deposit.Amount))

	syncedClaim, _ := suite.keeper.GetSavingsClaim(suite.ctx, claim.Owner)
	// indexes updated from global
	suite.Equal(globalIndexes, syncedClaim.RewardIndexes)
	// new reward is (new index - old index) * user shares
	suite.Equal(
		cs(c("rewarddenom", 1_000_001_000_000)).Add(originalReward...),
		syncedClaim.Reward,
	)
}

func (suite *SynchronizeSavingsRewardTests) TestClaimUnchangedWhenGlobalIndexesUnchanged() {
	denom := "test"
	unchangingIndexes := types.MultiRewardIndexes{
		{
			CollateralType: denom,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "rewarddenom",
					RewardFactor:   d("1000.001"),
				},
			},
		},
	}

	claim := types.SavingsClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner:  arbitraryAddress(),
			Reward: arbitraryCoins(),
		},
		RewardIndexes: unchangingIndexes,
	}
	suite.storeSavingsClaim(claim)

	suite.storeGlobalSavingsIndexes(unchangingIndexes)

	userShares := i(1e9)
	deposit := savingstypes.NewDeposit(claim.Owner, sdk.NewCoins(sdk.NewCoin(denom, userShares)))
	suite.keeper.SynchronizeSavingsReward(suite.ctx, deposit, getDenoms(deposit.Amount))

	syncedClaim, _ := suite.keeper.GetSavingsClaim(suite.ctx, claim.Owner)
	// claim should have the same rewards and indexes as before
	suite.Equal(claim, syncedClaim)
}

func (suite *SynchronizeSavingsRewardTests) TestClaimUpdatedWhenNewRewardAdded() {
	originalReward := arbitraryCoins()
	newlyRewardedDenom := "newlyRewardedDenom"

	claim := types.SavingsClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner:  arbitraryAddress(),
			Reward: originalReward,
		},
		RewardIndexes: types.MultiRewardIndexes{
			{
				CollateralType: "currentlyRewardedDenom",
				RewardIndexes: types.RewardIndexes{
					{
						CollateralType: "reward",
						RewardFactor:   d("1000.001"),
					},
				},
			},
		},
	}
	suite.storeSavingsClaim(claim)

	globalIndexes := types.MultiRewardIndexes{
		{
			CollateralType: "currentlyRewardedDenom",
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "reward",
					RewardFactor:   d("2000.002"),
				},
			},
		},
		{
			CollateralType: newlyRewardedDenom,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "otherreward",
					// Indexes start at 0 when the reward is added by gov,
					// so this represents the syncing happening some time later.
					RewardFactor: d("1000.001"),
				},
			},
		},
	}
	suite.storeGlobalSwapIndexes(globalIndexes)

	userShares := i(1e9)
	deposit := savingstypes.NewDeposit(claim.Owner,
		sdk.NewCoins(
			sdk.NewCoin("currentlyRewardedDenom", userShares),
			sdk.NewCoin(newlyRewardedDenom, userShares),
		),
	)

	suite.keeper.SynchronizeSavingsReward(suite.ctx, deposit, []string{newlyRewardedDenom})

	syncedClaim, _ := suite.keeper.GetSavingsClaim(suite.ctx, claim.Owner)
	// the new indexes should be added to the claim, but the old ones should be unchanged
	newlyRewardedIndexes, _ := globalIndexes.Get(newlyRewardedDenom)
	expectedIndexes := claim.RewardIndexes.With(newlyRewardedDenom, newlyRewardedIndexes)
	suite.Equal(expectedIndexes, syncedClaim.RewardIndexes)
	// new reward is (new index - old index) * shares for the synced pool
	// The old index for `newlyrewarded` isn't in the claim, so it's added starting at 0 for calculating the reward.
	suite.Equal(
		cs(c("otherreward", 1_000_001_000_000)).Add(originalReward...),
		syncedClaim.Reward,
	)
}

func (suite *SynchronizeSavingsRewardTests) TestClaimUnchangedWhenNoReward() {
	// When a pool is not rewarded but the user has deposited to that pool, and the claim is synced;
	// Then the claim should be the same.

	claim := types.SavingsClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner:  arbitraryAddress(),
			Reward: arbitraryCoins(),
		},
		RewardIndexes: nonEmptyMultiRewardIndexes,
	}
	suite.storeSavingsClaim(claim)

	denom := "nonRewardDenom"
	// No global indexes stored as this pool is not rewarded

	userShares := i(1e9)
	deposit := savingstypes.NewDeposit(claim.Owner, sdk.NewCoins(sdk.NewCoin(denom, userShares)))
	suite.keeper.SynchronizeSavingsReward(suite.ctx, deposit, getDenoms(deposit.Amount))

	syncedClaim, _ := suite.keeper.GetSwapClaim(suite.ctx, claim.Owner)
	suite.Equal(claim, syncedClaim)
}

func (suite *SynchronizeSavingsRewardTests) TestClaimUpdatedWhenNewRewardDenomAdded() {
	// When a new reward coin is added (via gov) to an already rewarded pool (that the user has already deposited to), and the claim is synced;
	// Then the user earns rewards for the time since the reward was added, and the new indexes are added.

	originalReward := arbitraryCoins()
	denom := "base"

	claim := types.SavingsClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner:  arbitraryAddress(),
			Reward: originalReward,
		},
		RewardIndexes: types.MultiRewardIndexes{
			{
				CollateralType: denom,
				RewardIndexes: types.RewardIndexes{
					{
						CollateralType: "reward",
						RewardFactor:   d("1000.001"),
					},
				},
			},
		},
	}
	suite.storeSavingsClaim(claim)

	globalIndexes := types.MultiRewardIndexes{
		{
			CollateralType: denom,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "reward",
					RewardFactor:   d("2000.002"),
				},
				{
					CollateralType: "otherreward",
					// Indexes start at 0 when the reward is added by gov,
					// so this represents the syncing happening some time later.
					RewardFactor: d("1000.001"),
				},
			},
		},
	}
	suite.storeGlobalSavingsIndexes(globalIndexes)

	userShares := i(1e9)
	deposit := savingstypes.NewDeposit(claim.Owner, sdk.NewCoins(sdk.NewCoin(denom, userShares)))
	suite.keeper.SynchronizeSavingsReward(suite.ctx, deposit, []string{denom})

	syncedClaim, _ := suite.keeper.GetSavingsClaim(suite.ctx, claim.Owner)
	// indexes should have the new reward denom added
	suite.Equal(globalIndexes, syncedClaim.RewardIndexes)
	// new reward is (new index - old index) * shares
	// The old index for `otherreward` isn't in the claim, so it's added starting at 0 for calculating the reward.
	suite.Equal(
		cs(c("reward", 1_000_001_000_000), c("otherreward", 1_000_001_000_000)).Add(originalReward...),
		syncedClaim.Reward,
	)
}

// func (suite *SynchronizeSavingsRewardTests) TestClaimUpdatedWhenGlobalIndexesIncreasedAndSourceIsZero() {
// 	// Given some time has passed (meaning the global indexes have increased)
// 	// When the claim is synced, but the user has no shares
// 	// The user earns no rewards for the time passed, but the claim indexes are updated

// 	poolID := "base"

// 	claim := types.SavingsClaim{
// 		BaseMultiClaim: types.BaseMultiClaim{
// 			Owner:  arbitraryAddress(),
// 			Reward: arbitraryCoins(),
// 		},
// 		RewardIndexes: types.MultiRewardIndexes{
// 			{
// 				CollateralType: poolID,
// 				RewardIndexes: types.RewardIndexes{
// 					{
// 						CollateralType: "rewarddenom",
// 						RewardFactor:   d("1000.001"),
// 					},
// 				},
// 			},
// 		},
// 	}
// 	suite.storeSavingsClaim(claim)

// 	globalIndexes := types.MultiRewardIndexes{
// 		{
// 			CollateralType: poolID,
// 			RewardIndexes: types.RewardIndexes{
// 				{
// 					CollateralType: "rewarddenom",
// 					RewardFactor:   d("2000.002"),
// 				},
// 			},
// 		},
// 	}
// 	suite.storeGlobalSavingsIndexes(globalIndexes)

// 	userShares := i(0)

// 	suite.keeper.SynchronizeSavingsReward(suite.ctx, poolID, claim.Owner, userShares)

// 	syncedClaim, _ := suite.keeper.GetSavingsClaim(suite.ctx, claim.Owner)
// 	// indexes updated from global
// 	suite.Equal(globalIndexes, syncedClaim.RewardIndexes)
// 	// reward is unchanged
// 	suite.Equal(claim.Reward, syncedClaim.Reward)
// }

// SavingsDepositBuilder is a tool for creating a savings deposit in tests.
// The builder inherits from savings.Deposit, so fields can be accessed directly if a helper method doesn't exist.
type SavingsDepositBuilder struct {
	savingstypes.Deposit
}

// NewSavingsDepositBuilder creates a SavingsDepositBuilder containing an empty deposit.
func NewSavingsDepositBuilder(depositor sdk.AccAddress) SavingsDepositBuilder {
	return SavingsDepositBuilder{
		Deposit: savingstypes.Deposit{
			Depositor: depositor,
		}}
}

// Build assembles and returns the final deposit.
func (builder SavingsDepositBuilder) Build() savingstypes.Deposit { return builder.Deposit }

// // WithSourceShares adds a deposit amount and factor such that the source shares for this deposit is equal to specified.
// // With a factor of 1, the deposit amount is the source shares. This picks an arbitrary factor to ensure factors are accounted for in production code.
// func (builder SavingsDepositBuilder) WithSourceShares(denom string, shares int64) SavingsDepositBuilder {
// 	if !builder.Amount.AmountOf(denom).Equal(sdk.ZeroInt()) {
// 		panic("adding to amount with existing denom not implemented")
// 	}
// 	if _, f := builder.Index.GetInterestFactor(denom); f {
// 		panic("adding to indexes with existing denom not implemented")
// 	}

// 	// pick arbitrary factor
// 	factor := sdk.MustNewDecFromStr("2")

// 	// Calculate deposit amount that would equal the requested source shares given the above factor.
// 	amt := sdk.NewInt(shares).Mul(factor.RoundInt())

// 	builder.Amount = builder.Amount.Add(sdk.NewCoin(denom, amt))
// 	builder.Index = builder.Index.SetInterestFactor(denom, factor)
// 	return builder
// }

// // WithArbitrarySourceShares adds arbitrary deposit amounts and indexes for each specified denom.
// func (builder SavingsDepositBuilder) WithArbitrarySourceShares(denoms ...string) SavingsDepositBuilder {
// 	const arbitraryShares = 1e9
// 	for _, denom := range denoms {
// 		builder = builder.WithSourceShares(denom, arbitraryShares)
// 	}
// 	return builder
// }

func getDenoms(coins sdk.Coins) []string {
	denoms := []string{}
	for _, coin := range coins {
		denoms = append(denoms, coin.Denom)
	}
	return denoms
}
