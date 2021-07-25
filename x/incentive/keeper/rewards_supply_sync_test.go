package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	hardtypes "github.com/kava-labs/kava/x/hard/types"
	"github.com/kava-labs/kava/x/incentive/types"
)

// SynchronizeHardSupplyRewardTests runs unit tests for the keeper.SynchronizeHardSupplyReward method
type SynchronizeHardSupplyRewardTests struct {
	unitTester
}

func TestSynchronizeHardSupplyReward(t *testing.T) {
	suite.Run(t, new(SynchronizeHardSupplyRewardTests))
}

func (suite *SynchronizeHardSupplyRewardTests) TestClaimIndexesAreUpdatedWhenGlobalIndexesHaveIncreased() {
	// This is the normal case

	claim := types.HardLiquidityProviderClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner: arbitraryAddress(),
		},
		SupplyRewardIndexes: nonEmptyMultiRewardIndexes,
	}
	suite.storeHardClaim(claim)

	globalIndexes := increaseAllRewardFactors(nonEmptyMultiRewardIndexes)
	suite.storeGlobalSupplyIndexes(globalIndexes)
	deposit := NewDepositBuilder(claim.Owner).
		WithArbitrarySourceShares(extractCollateralTypes(claim.SupplyRewardIndexes)...).
		Build()

	suite.keeper.SynchronizeHardSupplyReward(suite.ctx, deposit)

	syncedClaim, _ := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, claim.Owner)
	suite.Equal(globalIndexes, syncedClaim.SupplyRewardIndexes)
}
func (suite *SynchronizeHardSupplyRewardTests) TestClaimIndexesAreUnchangedWhenGlobalIndexesUnchanged() {
	// It should be safe to call SynchronizeHardSupplyReward multiple times

	unchangingIndexes := nonEmptyMultiRewardIndexes

	claim := types.HardLiquidityProviderClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner: arbitraryAddress(),
		},
		SupplyRewardIndexes: unchangingIndexes,
	}
	suite.storeHardClaim(claim)

	suite.storeGlobalSupplyIndexes(unchangingIndexes)

	deposit := NewDepositBuilder(claim.Owner).
		WithArbitrarySourceShares(extractCollateralTypes(unchangingIndexes)...).
		Build()

	suite.keeper.SynchronizeHardSupplyReward(suite.ctx, deposit)

	syncedClaim, _ := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, claim.Owner)
	suite.Equal(unchangingIndexes, syncedClaim.SupplyRewardIndexes)
}
func (suite *SynchronizeHardSupplyRewardTests) TestClaimIndexesAreUpdatedWhenNewRewardAdded() {
	// When a new reward is added (via gov) for a hard deposit denom the user has already deposited, and the claim is synced;
	// Then the new reward's index should be added to the claim.

	claim := types.HardLiquidityProviderClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner: arbitraryAddress(),
		},
		SupplyRewardIndexes: nonEmptyMultiRewardIndexes,
	}
	suite.storeHardClaim(claim)

	globalIndexes := appendUniqueMultiRewardIndex(nonEmptyMultiRewardIndexes)
	suite.storeGlobalSupplyIndexes(globalIndexes)

	deposit := NewDepositBuilder(claim.Owner).
		WithArbitrarySourceShares(extractCollateralTypes(globalIndexes)...).
		Build()

	suite.keeper.SynchronizeHardSupplyReward(suite.ctx, deposit)

	syncedClaim, _ := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, claim.Owner)
	suite.Equal(globalIndexes, syncedClaim.SupplyRewardIndexes)
}

func (suite *SynchronizeHardSupplyRewardTests) TestClaimIndexesAreUpdatedWhenNewRewardDenomAdded() {
	// When a new reward coin is added (via gov) to an already rewarded deposit denom (that the user has already deposited), and the claim is synced;
	// Then the new reward coin's index should be added to the claim.

	claim := types.HardLiquidityProviderClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner: arbitraryAddress(),
		},
		SupplyRewardIndexes: nonEmptyMultiRewardIndexes,
	}
	suite.storeHardClaim(claim)

	globalIndexes := appendUniqueRewardIndexToFirstItem(nonEmptyMultiRewardIndexes)
	suite.storeGlobalSupplyIndexes(globalIndexes)

	deposit := NewDepositBuilder(claim.Owner).
		WithArbitrarySourceShares(extractCollateralTypes(globalIndexes)...).
		Build()

	suite.keeper.SynchronizeHardSupplyReward(suite.ctx, deposit)

	syncedClaim, _ := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, claim.Owner)
	suite.Equal(globalIndexes, syncedClaim.SupplyRewardIndexes)
}

func (suite *SynchronizeHardSupplyRewardTests) TestRewardIsIncrementedWhenGlobalIndexesHaveIncreased() {
	// This is the normal case
	// Given some time has passed (meaning the global indexes have increased)
	// When the claim is synced
	// The user earns rewards for the time passed

	originalReward := arbitraryCoins()

	claim := types.HardLiquidityProviderClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner:  arbitraryAddress(),
			Reward: originalReward,
		},
		SupplyRewardIndexes: types.MultiRewardIndexes{
			{
				CollateralType: "depositdenom",
				RewardIndexes: types.RewardIndexes{
					{
						CollateralType: "rewarddenom",
						RewardFactor:   d("1000.001"),
					},
				},
			},
		},
	}
	suite.storeHardClaim(claim)

	suite.storeGlobalSupplyIndexes(types.MultiRewardIndexes{
		{
			CollateralType: "depositdenom",
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "rewarddenom",
					RewardFactor:   d("2000.002"),
				},
			},
		},
	})

	deposit := NewDepositBuilder(claim.Owner).
		WithSourceShares("depositdenom", 1e9).
		Build()

	suite.keeper.SynchronizeHardSupplyReward(suite.ctx, deposit)

	// new reward is (new index - old index) * deposit amount
	syncedClaim, _ := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, claim.Owner)
	suite.Equal(
		cs(c("rewarddenom", 1_000_001_000_000)).Add(originalReward...),
		syncedClaim.Reward,
	)
}

func (suite *SynchronizeHardSupplyRewardTests) TestRewardIsIncrementedWhenNewRewardAdded() {
	// When a new reward is added (via gov) for a hard deposit denom the user has already deposited, and the claim is synced
	// Then the user earns rewards for the time since the reward was added

	originalReward := arbitraryCoins()
	claim := types.HardLiquidityProviderClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner:  arbitraryAddress(),
			Reward: originalReward,
		},
		SupplyRewardIndexes: types.MultiRewardIndexes{
			{
				CollateralType: "rewarded",
				RewardIndexes: types.RewardIndexes{
					{
						CollateralType: "reward",
						RewardFactor:   d("1000.001"),
					},
				},
			},
		},
	}
	suite.storeHardClaim(claim)

	globalIndexes := types.MultiRewardIndexes{
		{
			CollateralType: "rewarded",
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "reward",
					RewardFactor:   d("2000.002"),
				},
			},
		},
		{
			CollateralType: "newlyrewarded",
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
	suite.storeGlobalSupplyIndexes(globalIndexes)

	deposit := NewDepositBuilder(claim.Owner).
		WithSourceShares("rewarded", 1e9).
		WithSourceShares("newlyrewarded", 1e9).
		Build()

	suite.keeper.SynchronizeHardSupplyReward(suite.ctx, deposit)

	// new reward is (new index - old index) * deposit amount for each deposited denom
	// The old index for `newlyrewarded` isn't in the claim, so it's added starting at 0 for calculating the reward.
	syncedClaim, _ := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, claim.Owner)
	suite.Equal(
		cs(c("otherreward", 1_000_001_000_000), c("reward", 1_000_001_000_000)).Add(originalReward...),
		syncedClaim.Reward,
	)
}
func (suite *SynchronizeHardSupplyRewardTests) TestRewardIsIncrementedWhenNewRewardDenomAdded() {
	// When a new reward coin is added (via gov) to an already rewarded deposit denom (that the user has already deposited), and the claim is synced;
	// Then the user earns rewards for the time since the reward was added

	originalReward := arbitraryCoins()
	claim := types.HardLiquidityProviderClaim{
		BaseMultiClaim: types.BaseMultiClaim{
			Owner:  arbitraryAddress(),
			Reward: originalReward,
		},
		SupplyRewardIndexes: types.MultiRewardIndexes{
			{
				CollateralType: "deposited",
				RewardIndexes: types.RewardIndexes{
					{
						CollateralType: "reward",
						RewardFactor:   d("1000.001"),
					},
				},
			},
		},
	}
	suite.storeHardClaim(claim)

	globalIndexes := types.MultiRewardIndexes{
		{
			CollateralType: "deposited",
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
	suite.storeGlobalSupplyIndexes(globalIndexes)

	deposit := NewDepositBuilder(claim.Owner).
		WithSourceShares("deposited", 1e9).
		Build()

	suite.keeper.SynchronizeHardSupplyReward(suite.ctx, deposit)

	// new reward is (new index - old index) * deposit amount for each deposited denom
	// The old index for `otherreward` isn't in the claim, so it's added starting at 0 for calculating the reward.
	syncedClaim, _ := suite.keeper.GetHardLiquidityProviderClaim(suite.ctx, claim.Owner)
	suite.Equal(
		cs(c("reward", 1_000_001_000_000), c("otherreward", 1_000_001_000_000)).Add(originalReward...),
		syncedClaim.Reward,
	)
}

// DepositBuilder is a tool for creating a hard deposit in tests.
// The builder inherits from hard.Deposit, so fields can be accessed directly if a helper method doesn't exist.
type DepositBuilder struct {
	hardtypes.Deposit
}

// NewDepositBuilder creates a DepositBuilder containing an empty deposit.
func NewDepositBuilder(depositor sdk.AccAddress) DepositBuilder {
	return DepositBuilder{
		Deposit: hardtypes.Deposit{
			Depositor: depositor,
		}}
}

// Build assembles and returns the final deposit.
func (builder DepositBuilder) Build() hardtypes.Deposit { return builder.Deposit }

// WithSourceShares adds a deposit amount and factor such that the source shares for this deposit is equal to specified.
// With a factor of 1, the deposit amount is the source shares. This picks an arbitrary factor to ensure factors are accounted for in production code.
func (builder DepositBuilder) WithSourceShares(denom string, shares int64) DepositBuilder {
	if !builder.Amount.AmountOf(denom).Equal(sdk.ZeroInt()) {
		panic("adding to amount with existing denom not implemented")
	}
	if _, f := builder.Index.GetInterestFactor(denom); f {
		panic("adding to indexes with existing denom not implemented")
	}

	// pick arbitrary factor
	// factor := sdk.MustNewDecFromStr("2")

	// Calculate deposit amount that would equal the requested source shares given the above factor.
	amt := sdk.NewInt(shares) //.Mul(factor.RoundInt())

	builder.Amount = builder.Amount.Add(sdk.NewCoin(denom, amt))
	// builder.Index = builder.Index.SetInterestFactor(denom, factor)
	return builder
}

// WithArbitrarySourceShares adds arbitrary deposit amounts and indexes for each specified denom.
func (builder DepositBuilder) WithArbitrarySourceShares(denoms ...string) DepositBuilder {
	const arbitraryShares = 1e9
	for _, denom := range denoms {
		builder = builder.WithSourceShares(denom, arbitraryShares)
	}
	return builder
}
