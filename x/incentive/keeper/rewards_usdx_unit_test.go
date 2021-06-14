package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	"github.com/kava-labs/kava/x/incentive/types"
)

// usdxRewardsUnitTester contains common methods for running unit tests for keeper methods related to the USDX minting rewards
type usdxRewardsUnitTester struct {
	unitTester
}

func (suite *usdxRewardsUnitTester) storeGlobalUSDXIndexes(indexes types.RewardIndexes) {
	for _, ri := range indexes {
		suite.keeper.SetUSDXMintingRewardFactor(suite.ctx, ri.CollateralType, ri.RewardFactor)
	}
}
func (suite *usdxRewardsUnitTester) storeClaim(claim types.USDXMintingClaim) {
	suite.keeper.SetUSDXMintingClaim(suite.ctx, claim)
}

type InitializeUSDXMintingClaimTests struct {
	usdxRewardsUnitTester
}

func TestInitializeUSDXMintingClaims(t *testing.T) {
	suite.Run(t, new(InitializeUSDXMintingClaimTests))
}

func (suite *InitializeUSDXMintingClaimTests) TestClaimIndexIsSetWhenClaimDoesNotExist() {
	collateralType := "bnb-a"

	subspace := paramsWithSingleUSDXRewardPeriod(collateralType)
	suite.keeper = suite.NewKeeper(subspace, nil, nil, nil, nil, nil)

	cdp := NewCDPBuilder(arbitraryAddress(), collateralType).Build()

	globalIndexes := types.RewardIndexes{{
		CollateralType: collateralType,
		RewardFactor:   d("0.2"),
	}}
	suite.storeGlobalUSDXIndexes(globalIndexes)

	suite.keeper.InitializeUSDXMintingClaim(suite.ctx, cdp)

	syncedClaim, f := suite.keeper.GetUSDXMintingClaim(suite.ctx, cdp.Owner)
	suite.True(f)
	suite.Equal(globalIndexes, syncedClaim.RewardIndexes)
}

func (suite *InitializeUSDXMintingClaimTests) TestClaimIndexIsSetWhenClaimExists() {
	collateralType := "bnb-a"

	subspace := paramsWithSingleUSDXRewardPeriod(collateralType)
	suite.keeper = suite.NewKeeper(subspace, nil, nil, nil, nil, nil)

	claim := types.USDXMintingClaim{
		BaseClaim: types.BaseClaim{
			Owner: arbitraryAddress(),
		},
		RewardIndexes: types.RewardIndexes{{
			CollateralType: collateralType,
			RewardFactor:   d("0.1"),
		}},
	}
	suite.storeClaim(claim)

	globalIndexes := types.RewardIndexes{{
		CollateralType: collateralType,
		RewardFactor:   d("0.2"),
	}}
	suite.storeGlobalUSDXIndexes(globalIndexes)

	cdp := NewCDPBuilder(claim.Owner, collateralType).Build()

	suite.keeper.InitializeUSDXMintingClaim(suite.ctx, cdp)

	syncedClaim, _ := suite.keeper.GetUSDXMintingClaim(suite.ctx, cdp.Owner)
	suite.Equal(globalIndexes, syncedClaim.RewardIndexes)
}

type SynchronizeUSDXMintingRewardTests struct {
	usdxRewardsUnitTester
}

func TestSynchronizeUSDXMintingReward(t *testing.T) {
	suite.Run(t, new(SynchronizeUSDXMintingRewardTests))
}
func (suite *SynchronizeUSDXMintingRewardTests) TestRewardUnchangedWhenGlobalIndexesUnchanged() {
	unchangingRewardIndexes := nonEmptyRewardIndexes
	collateralType := extractFirstCollateralType(unchangingRewardIndexes)

	subspace := paramsWithSingleUSDXRewardPeriod(collateralType)
	suite.keeper = suite.NewKeeper(subspace, nil, nil, nil, nil, nil)

	claim := types.USDXMintingClaim{
		BaseClaim: types.BaseClaim{
			Owner:  arbitraryAddress(),
			Reward: c(types.USDXMintingRewardDenom, 0),
		},
		RewardIndexes: unchangingRewardIndexes,
	}
	suite.storeClaim(claim)

	suite.storeGlobalUSDXIndexes(unchangingRewardIndexes)

	cdp := NewCDPBuilder(claim.Owner, collateralType).WithPrincipal(i(1e12)).Build()

	suite.keeper.SynchronizeUSDXMintingReward(suite.ctx, cdp)

	syncedClaim, _ := suite.keeper.GetUSDXMintingClaim(suite.ctx, claim.Owner)
	suite.Equal(claim.Reward, syncedClaim.Reward)
}

func (suite *SynchronizeUSDXMintingRewardTests) TestRewardIsIncrementedWhenGlobalIndexIncreased() {
	collateralType := "bnb-a"

	subspace := paramsWithSingleUSDXRewardPeriod(collateralType)
	suite.keeper = suite.NewKeeper(subspace, nil, nil, nil, nil, nil)

	claim := types.USDXMintingClaim{
		BaseClaim: types.BaseClaim{
			Owner:  arbitraryAddress(),
			Reward: c(types.USDXMintingRewardDenom, 0),
		},
		RewardIndexes: types.RewardIndexes{
			{
				CollateralType: collateralType,
				RewardFactor:   d("0.1"),
			},
		},
	}
	suite.storeClaim(claim)

	globalIndexes := types.RewardIndexes{
		{
			CollateralType: collateralType,
			RewardFactor:   d("0.2"),
		},
	}
	suite.storeGlobalUSDXIndexes(globalIndexes)

	cdp := NewCDPBuilder(claim.Owner, collateralType).WithPrincipal(i(1e12)).Build()

	suite.keeper.SynchronizeUSDXMintingReward(suite.ctx, cdp)

	syncedClaim, _ := suite.keeper.GetUSDXMintingClaim(suite.ctx, claim.Owner)
	// reward is ( new index - old index ) * cdp.TotalPrincipal
	suite.Equal(c(types.USDXMintingRewardDenom, 1e11), syncedClaim.Reward)
}

func (suite *SynchronizeUSDXMintingRewardTests) TestRewardIsIncrementedWhenNewRewardAddedAndClaimDoesNotExit() {
	collateralType := "bnb-a"

	subspace := paramsWithSingleUSDXRewardPeriod(collateralType)
	suite.keeper = suite.NewKeeper(subspace, nil, nil, nil, nil, nil)

	globalIndexes := types.RewardIndexes{
		{
			CollateralType: collateralType,
			RewardFactor:   d("0.2"),
		},
	}
	suite.storeGlobalUSDXIndexes(globalIndexes)

	cdp := NewCDPBuilder(arbitraryAddress(), collateralType).WithPrincipal(i(1e12)).Build()

	suite.keeper.SynchronizeUSDXMintingReward(suite.ctx, cdp)

	syncedClaim, _ := suite.keeper.GetUSDXMintingClaim(suite.ctx, cdp.Owner)
	// The global index was not around when this cdp was created as it was not stored in a claim.
	// Therefore it must have been added via params after.
	// To include rewards since the params were updated, the old index should be assumed to be 0.
	// reward is ( new index - old index ) * cdp.TotalPrincipal
	suite.Equal(c(types.USDXMintingRewardDenom, 2e11), syncedClaim.Reward)
}
func (suite *SynchronizeUSDXMintingRewardTests) TestClaimIndexIsUpdatedWhenGlobalIndexIncreased() {
	claimsRewardIndexes := nonEmptyRewardIndexes
	collateralType := extractFirstCollateralType(claimsRewardIndexes)

	subspace := paramsWithSingleUSDXRewardPeriod(collateralType)
	suite.keeper = suite.NewKeeper(subspace, nil, nil, nil, nil, nil)

	claim := types.USDXMintingClaim{
		BaseClaim: types.BaseClaim{
			Owner:  arbitraryAddress(),
			Reward: c(types.USDXMintingRewardDenom, 0),
		},
		RewardIndexes: claimsRewardIndexes,
	}
	suite.storeClaim(claim)

	globalIndexes := increaseRewardFactors(claimsRewardIndexes)
	suite.storeGlobalUSDXIndexes(globalIndexes)

	cdp := NewCDPBuilder(claim.Owner, collateralType).Build()

	suite.keeper.SynchronizeUSDXMintingReward(suite.ctx, cdp)

	syncedClaim, _ := suite.keeper.GetUSDXMintingClaim(suite.ctx, claim.Owner)

	// Only the claim's index for `collateralType` should have been changed
	i, _ := globalIndexes.Get(collateralType)
	expectedIndexes := claimsRewardIndexes.With(collateralType, i)
	suite.Equal(expectedIndexes, syncedClaim.RewardIndexes)
}

func (suite *SynchronizeUSDXMintingRewardTests) TestClaimIndexIsUpdatedWhenNewRewardAddedAndClaimAlreadyExists() {
	claimsRewardIndexes := types.RewardIndexes{
		{
			CollateralType: "bnb-a",
			RewardFactor:   d("0.1"),
		},
		{
			CollateralType: "busd-b",
			RewardFactor:   d("0.4"),
		},
	}
	newRewardIndex := types.NewRewardIndex("xrp-a", d("0.0001"))

	subspace := paramsWithSingleUSDXRewardPeriod(newRewardIndex.CollateralType)
	suite.keeper = suite.NewKeeper(subspace, nil, nil, nil, nil, nil)

	claim := types.USDXMintingClaim{
		BaseClaim: types.BaseClaim{
			Owner:  arbitraryAddress(),
			Reward: c(types.USDXMintingRewardDenom, 0),
		},
		RewardIndexes: claimsRewardIndexes,
	}
	suite.storeClaim(claim)

	globalIndexes := increaseRewardFactors(claimsRewardIndexes)
	globalIndexes = append(globalIndexes, newRewardIndex)
	suite.storeGlobalUSDXIndexes(globalIndexes)

	cdp := NewCDPBuilder(claim.Owner, newRewardIndex.CollateralType).Build()

	suite.keeper.SynchronizeUSDXMintingReward(suite.ctx, cdp)

	syncedClaim, _ := suite.keeper.GetUSDXMintingClaim(suite.ctx, claim.Owner)

	// Only the claim's index for `collateralType` should have been changed
	expectedIndexes := claimsRewardIndexes.With(newRewardIndex.CollateralType, newRewardIndex.RewardFactor)
	suite.Equal(expectedIndexes, syncedClaim.RewardIndexes)
}

type cdpBuilder struct {
	cdptypes.CDP
}

func NewCDPBuilder(owner sdk.AccAddress, collateralType string) cdpBuilder {
	return cdpBuilder{
		CDP: cdptypes.CDP{
			Owner: owner,
			Type:  collateralType,
			// The zero value of Principal and AccumulatedFees (type sdk.Coin) is invalid as the denom is ""
			// Set them to the default denom, but with 0 amount.
			Principal:       c(cdptypes.DefaultStableDenom, 0),
			AccumulatedFees: c(cdptypes.DefaultStableDenom, 0),
		}}
}

func (builder cdpBuilder) Build() cdptypes.CDP { return builder.CDP }

func (builder cdpBuilder) WithPrincipal(principal sdk.Int) cdpBuilder {
	builder.Principal = sdk.NewCoin(cdptypes.DefaultStableDenom, principal)
	return builder
}

var nonEmptyRewardIndexes = types.RewardIndexes{
	{
		CollateralType: "bnb-a",
		RewardFactor:   d("0.1"),
	},
	{
		CollateralType: "busd-b",
		RewardFactor:   d("0.4"),
	},
}

func paramsWithSingleUSDXRewardPeriod(collateralType string) types.ParamSubspace {
	return &fakeParamSubspace{
		params: types.Params{
			USDXMintingRewardPeriods: types.RewardPeriods{
				{
					CollateralType: collateralType,
				},
			},
		},
	}
}

func extractFirstCollateralType(indexes types.RewardIndexes) string {
	if len(indexes) == 0 {
		panic("cannot extract a collateral type from 0 length RewardIndexes")
	}
	return indexes[0].CollateralType
}
