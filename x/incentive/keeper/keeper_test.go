package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/incentive/keeper"
	"github.com/kava-labs/kava/x/incentive/types"
)

// Test suite used for all keeper tests
type KeeperTestSuite struct {
	suite.Suite

	keeper keeper.Keeper

	app app.TestApp
	ctx sdk.Context

	genesisTime time.Time
	addrs       []sdk.AccAddress
}

// SetupTest is run automatically before each suite test
func (suite *KeeperTestSuite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)

	_, suite.addrs = app.GeneratePrivKeyAddressPairs(5)

	suite.genesisTime = time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC)
}

func (suite *KeeperTestSuite) SetupApp() {
	suite.app = app.NewTestApp()

	suite.keeper = suite.app.GetIncentiveKeeper()

	suite.ctx = suite.app.NewContext(true, abci.Header{Height: 1, Time: suite.genesisTime})
}

func (suite *KeeperTestSuite) TestGetSetDeleteUSDXMintingClaim() {
	suite.SetupApp()
	c := types.NewUSDXMintingClaim(suite.addrs[0], c("ukava", 1000000), types.RewardIndexes{types.NewRewardIndex("bnb-a", sdk.ZeroDec())})
	_, found := suite.keeper.GetUSDXMintingClaim(suite.ctx, suite.addrs[0])
	suite.Require().False(found)
	suite.Require().NotPanics(func() {
		suite.keeper.SetUSDXMintingClaim(suite.ctx, c)
	})
	testC, found := suite.keeper.GetUSDXMintingClaim(suite.ctx, suite.addrs[0])
	suite.Require().True(found)
	suite.Require().Equal(c, testC)
	suite.Require().NotPanics(func() {
		suite.keeper.DeleteUSDXMintingClaim(suite.ctx, suite.addrs[0])
	})
	_, found = suite.keeper.GetUSDXMintingClaim(suite.ctx, suite.addrs[0])
	suite.Require().False(found)
}

func (suite *KeeperTestSuite) TestIterateUSDXMintingClaims() {
	suite.SetupApp()
	for i := 0; i < len(suite.addrs); i++ {
		c := types.NewUSDXMintingClaim(suite.addrs[i], c("ukava", 100000), types.RewardIndexes{types.NewRewardIndex("bnb-a", sdk.ZeroDec())})
		suite.Require().NotPanics(func() {
			suite.keeper.SetUSDXMintingClaim(suite.ctx, c)
		})
	}
	claims := types.USDXMintingClaims{}
	suite.keeper.IterateUSDXMintingClaims(suite.ctx, func(c types.USDXMintingClaim) bool {
		claims = append(claims, c)
		return false
	})
	suite.Require().Equal(len(suite.addrs), len(claims))

	claims = suite.keeper.GetAllUSDXMintingClaims(suite.ctx)
	suite.Require().Equal(len(suite.addrs), len(claims))
}

func (suite *KeeperTestSuite) TestGetSetSwapRewardIndexes() {
	testCases := []struct {
		name      string
		poolName  string
		indexes   types.RewardIndexes
		setPanics bool
	}{
		{
			name:     "two factors can be written and read",
			poolName: "btc/usdx",
			indexes: types.RewardIndexes{
				{
					CollateralType: "hard",
					RewardFactor:   d("0.02"),
				},
				{
					CollateralType: "ukava",
					RewardFactor:   d("0.04"),
				},
			},
		},
		{
			name:     "indexes with empty pool name can be written and read",
			poolName: "",
			indexes: types.RewardIndexes{
				{
					CollateralType: "hard",
					RewardFactor:   d("0.02"),
				},
				{
					CollateralType: "ukava",
					RewardFactor:   d("0.04"),
				},
			},
		},
		{
			name:      "setting empty indexes panics",
			poolName:  "btc/usdx",
			indexes:   types.RewardIndexes{},
			setPanics: true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupApp()

			_, found := suite.keeper.GetSwapRewardIndexes(suite.ctx, tc.poolName)
			suite.False(found)

			setFunc := func() { suite.keeper.SetSwapRewardIndexes(suite.ctx, tc.poolName, tc.indexes) }
			if tc.setPanics {
				suite.Panics(setFunc)
				return
			} else {
				suite.NotPanics(setFunc)
			}

			storedIndexes, found := suite.keeper.GetSwapRewardIndexes(suite.ctx, tc.poolName)
			suite.True(found)
			suite.Equal(tc.indexes, storedIndexes)
		})
	}
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
