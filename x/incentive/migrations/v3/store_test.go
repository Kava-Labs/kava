package v3_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/incentive/testutil"
	"github.com/kava-labs/kava/x/incentive/types"
	"github.com/stretchr/testify/suite"

	v3 "github.com/kava-labs/kava/x/incentive/migrations/v3"
)

type StoreMigrateTestSuite struct {
	testutil.IntegrationTester

	Addrs []sdk.AccAddress

	keeper   testutil.TestKeeper
	storeKey sdk.StoreKey
	cdc      codec.Codec
}

func TestStoreMigrateTestSuite(t *testing.T) {
	suite.Run(t, new(StoreMigrateTestSuite))
}

func (suite *StoreMigrateTestSuite) SetupTest() {
	suite.IntegrationTester.SetupTest()

	suite.keeper = testutil.TestKeeper{
		Keeper: suite.App.GetIncentiveKeeper(),
	}

	_, suite.Addrs = app.GeneratePrivKeyAddressPairs(5)
	suite.cdc = suite.App.AppCodec()
	suite.storeKey = suite.App.GetKeys()[types.StoreKey]

	suite.StartChain()
}

func (suite *StoreMigrateTestSuite) TestMigrateEarnClaims() {
	store := suite.Ctx.KVStore(suite.storeKey)

	// Create v2 earn claims
	claim1 := types.NewEarnClaim(
		suite.Addrs[0],
		sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(100))),
		types.MultiRewardIndexes{
			types.NewMultiRewardIndex("bnb-a", types.RewardIndexes{
				types.NewRewardIndex("bnb", sdk.NewDec(1)),
			}),
		},
	)

	claim2 := types.NewEarnClaim(
		suite.Addrs[1],
		sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(100))),
		types.MultiRewardIndexes{
			types.NewMultiRewardIndex("ukava", types.RewardIndexes{
				types.NewRewardIndex("ukava", sdk.NewDec(1)),
			}),
		},
	)

	suite.keeper.SetEarnClaim(suite.Ctx, claim1)
	suite.keeper.SetEarnClaim(suite.Ctx, claim2)

	// Run earn claim migrations
	err := v3.MigrateEarnClaims(store, suite.cdc)
	suite.Require().NoError(err)

	// Check that the claim was migrated correctly
	newClaim1, found := suite.keeper.Store.GetClaim(suite.Ctx, types.CLAIM_TYPE_EARN, claim1.Owner)
	suite.Require().True(found)
	suite.Require().Equal(claim1.Owner, newClaim1.Owner)

	newClaim2, found := suite.keeper.Store.GetClaim(suite.Ctx, types.CLAIM_TYPE_EARN, claim2.Owner)
	suite.Require().True(found)
	suite.Require().Equal(claim2.Owner, newClaim2.Owner)

	// Ensure removed from old store
	_, found = suite.keeper.GetEarnClaim(suite.Ctx, claim1.Owner)
	suite.Require().False(found)

	_, found = suite.keeper.GetEarnClaim(suite.Ctx, claim2.Owner)
	suite.Require().False(found)
}

func (suite *StoreMigrateTestSuite) TestMigrateHardClaims() {
	store := suite.Ctx.KVStore(suite.storeKey)

	// Create v2 earn claims
	claim1 := types.NewHardLiquidityProviderClaim(
		suite.Addrs[0],
		sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(100))),
		types.MultiRewardIndexes{
			types.NewMultiRewardIndex("bnb-a", types.RewardIndexes{
				types.NewRewardIndex("bnb", sdk.NewDec(1)),
			}),
		},
		types.MultiRewardIndexes{
			types.NewMultiRewardIndex("bnb-b", types.RewardIndexes{
				types.NewRewardIndex("bnb", sdk.NewDec(2)),
			}),
		},
	)

	claim2 := types.NewHardLiquidityProviderClaim(
		suite.Addrs[1],
		sdk.NewCoins(sdk.NewCoin("usdx", sdk.NewInt(100))),
		types.MultiRewardIndexes{
			types.NewMultiRewardIndex("ukava", types.RewardIndexes{
				types.NewRewardIndex("ukava", sdk.NewDec(1)),
			}),
		},
		// No borrows
		nil,
	)

	suite.keeper.SetHardLiquidityProviderClaim(suite.Ctx, claim1)
	suite.keeper.SetHardLiquidityProviderClaim(suite.Ctx, claim2)

	// Run earn claim migrations
	err := v3.MigrateHardClaims(store, suite.cdc)
	suite.Require().NoError(err)

	// Check that the claim was migrated correctly
	// Two claims, one for supply and one for borrow
	newClaim1Supply, found := suite.keeper.Store.GetClaim(suite.Ctx, types.CLAIM_TYPE_HARD_SUPPLY, claim1.Owner)
	suite.Require().True(found)
	suite.Require().Equal(claim1.Owner, newClaim1Supply.Owner)
	suite.Require().Equal(claim1.SupplyRewardIndexes, newClaim1Supply.RewardIndexes)

	newClaim1Borrow, found := suite.keeper.Store.GetClaim(suite.Ctx, types.CLAIM_TYPE_HARD_BORROW, claim1.Owner)
	suite.Require().True(found)
	suite.Require().Equal(claim1.BorrowRewardIndexes, newClaim1Borrow.RewardIndexes)
	suite.Require().Equal(sdk.Coins(nil), newClaim1Borrow.Reward, "borrow claim should have no rewards")

	// Second claim has no borrows so only one supply claim
	newClaim2Supply, found := suite.keeper.Store.GetClaim(suite.Ctx, types.CLAIM_TYPE_HARD_SUPPLY, claim2.Owner)
	suite.Require().True(found)
	suite.Require().Equal(claim2.Owner, newClaim2Supply.Owner)
	suite.Require().Equal(claim2.SupplyRewardIndexes, newClaim2Supply.RewardIndexes)

	_, found = suite.keeper.Store.GetClaim(suite.Ctx, types.CLAIM_TYPE_HARD_BORROW, claim2.Owner)
	suite.Require().False(found, "borrow claim should not exist if old claim has no borrows")

	// Ensure removed from old store
	_, found = suite.keeper.GetHardLiquidityProviderClaim(suite.Ctx, claim1.Owner)
	suite.Require().False(found)

	_, found = suite.keeper.GetHardLiquidityProviderClaim(suite.Ctx, claim2.Owner)
	suite.Require().False(found)
}

func (suite *StoreMigrateTestSuite) TestMigrateAccrualTimes() {
	store := suite.Ctx.KVStore(suite.storeKey)
	vaultDenom1 := "ukava"
	vaultDenom2 := "usdc"

	// Create v2 accrual times
	accrualTime1 := time.Now()
	accrualTime2 := time.Now().Add(time.Hour * 24)
	suite.keeper.SetEarnRewardAccrualTime(suite.Ctx, vaultDenom1, accrualTime1)
	suite.keeper.SetEarnRewardAccrualTime(suite.Ctx, vaultDenom2, accrualTime2)

	// Run accrual time migrations
	err := v3.MigrateAccrualTimes(store, suite.cdc, types.CLAIM_TYPE_EARN)
	suite.Require().NoError(err)

	// Check that the accrual time was migrated correctly
	newAccrualTime1, found := suite.keeper.Store.GetRewardAccrualTime(suite.Ctx, types.CLAIM_TYPE_EARN, vaultDenom1)
	suite.Require().True(found)
	suite.Require().Equal(accrualTime1.Unix(), newAccrualTime1.Unix())

	newAccrualTime2, found := suite.keeper.Store.GetRewardAccrualTime(suite.Ctx, types.CLAIM_TYPE_EARN, vaultDenom2)
	suite.Require().True(found)
	suite.Require().Equal(accrualTime2.Unix(), newAccrualTime2.Unix())

	// Ensure removed from old store
	_, found = suite.keeper.GetEarnRewardAccrualTime(suite.Ctx, vaultDenom1)
	suite.Require().False(found)
	_, found = suite.keeper.GetEarnRewardAccrualTime(suite.Ctx, vaultDenom2)
	suite.Require().False(found)
}

func (suite *StoreMigrateTestSuite) TestMigrateRewardIndexes() {
	store := suite.Ctx.KVStore(suite.storeKey)
	vaultDenom1 := "ukava"
	vaultDenom2 := "usdc"

	rewardIndexes1 := types.RewardIndexes{
		types.NewRewardIndex("ukava", sdk.NewDec(1)),
		types.NewRewardIndex("hard", sdk.NewDec(2)),
	}
	rewardIndexes2 := types.RewardIndexes{
		types.NewRewardIndex("ukava", sdk.NewDec(4)),
		types.NewRewardIndex("swp", sdk.NewDec(10)),
	}

	suite.keeper.SetEarnRewardIndexes(suite.Ctx, vaultDenom1, rewardIndexes1)
	suite.keeper.SetEarnRewardIndexes(suite.Ctx, vaultDenom2, rewardIndexes2)

	err := v3.MigrateRewardIndexes(store, suite.cdc, types.CLAIM_TYPE_EARN)
	suite.Require().NoError(err)

	newRewardIndexes1, found := suite.keeper.Store.GetRewardIndexesOfClaimType(suite.Ctx, types.CLAIM_TYPE_EARN, vaultDenom1)
	suite.Require().True(found)
	suite.Require().Equal(rewardIndexes1, newRewardIndexes1)

	newRewardIndexes2, found := suite.keeper.Store.GetRewardIndexesOfClaimType(suite.Ctx, types.CLAIM_TYPE_EARN, vaultDenom2)
	suite.Require().True(found)
	suite.Require().Equal(rewardIndexes2, newRewardIndexes2)

	// Ensure removed from old store
	_, found = suite.keeper.GetEarnRewardIndexes(suite.Ctx, vaultDenom1)
	suite.Require().False(found)

	_, found = suite.keeper.GetEarnRewardIndexes(suite.Ctx, vaultDenom2)
	suite.Require().False(found)
}
