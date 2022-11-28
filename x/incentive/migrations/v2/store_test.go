package v2_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/incentive/testutil"
	"github.com/kava-labs/kava/x/incentive/types"
	"github.com/stretchr/testify/suite"

	v2 "github.com/kava-labs/kava/x/incentive/migrations/v2"
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

	// Create v1 earn claims
	claim1 := types.NewEarnClaim(
		suite.Addrs[0],
		sdk.NewCoins(sdk.NewCoin("bnb", sdk.NewInt(100))),
		types.MultiRewardIndexes{
			types.NewMultiRewardIndex("bnb-a", types.RewardIndexes{
				types.NewRewardIndex("bnb", sdk.NewDec(1)),
			}),
		},
	)
	suite.keeper.SetEarnClaim(suite.Ctx, claim1)

	// Run earn claim migrations
	err := v2.MigrateEarnClaims(store, suite.cdc)
	suite.Require().NoError(err)

	// Check that the claim was migrated correctly
	claim, found := suite.keeper.Store.GetClaim(suite.Ctx, types.CLAIM_TYPE_EARN, claim1.Owner)
	suite.Require().True(found)
	suite.Require().Equal(claim1.Owner, claim.Owner)
}

func (suite *StoreMigrateTestSuite) TestMigrateAccrualTimes() {
	store := suite.Ctx.KVStore(suite.storeKey)
	vaultDenom := "ukava"

	// Create v1 accrual times
	accrualTime1 := time.Now().UTC()
	suite.keeper.SetEarnRewardAccrualTime(suite.Ctx, vaultDenom, accrualTime1)

	// Run accrual time migrations
	err := v2.MigrateAccrualTimes(store, suite.cdc, types.CLAIM_TYPE_EARN)
	suite.Require().NoError(err)

	// Check that the accrual time was migrated correctly
	accrualTime, found := suite.keeper.Store.GetRewardAccrualTime(suite.Ctx, types.CLAIM_TYPE_EARN, vaultDenom)
	suite.Require().True(found)
	suite.Require().Equal(accrualTime1, accrualTime)
}

func (suite *StoreMigrateTestSuite) TestMigrateRewardIndexes() {
	store := suite.Ctx.KVStore(suite.storeKey)
	vaultDenom := "ukava"

	rewardIndexes := types.RewardIndexes{
		types.NewRewardIndex("ukava", sdk.NewDec(1)),
		types.NewRewardIndex("hard", sdk.NewDec(2)),
	}
	suite.keeper.SetEarnRewardIndexes(suite.Ctx, vaultDenom, rewardIndexes)

	err := v2.MigrateRewardIndexes(store, suite.cdc, types.CLAIM_TYPE_EARN)
	suite.Require().NoError(err)

	rewardIndexesMigrated, found := suite.keeper.Store.GetRewardIndexesOfClaimType(suite.Ctx, types.CLAIM_TYPE_EARN, vaultDenom)
	suite.Require().True(found)
	suite.Require().Equal(rewardIndexes, rewardIndexesMigrated)
}
