package distribution_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	distrquery "github.com/kava-labs/kava/app/query/distribution"
	communitytypes "github.com/kava-labs/kava/x/community/types"
)

type queryTestSuite struct {
	suite.Suite

	App app.TestApp
	Ctx sdk.Context

	queryClient          distrtypes.QueryClient
	communityPoolAddress sdk.AccAddress
}

func (suite *queryTestSuite) SetupTest() {
	app.SetSDKConfig()
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})

	tApp.InitializeFromGenesisStates()

	suite.App = tApp
	suite.Ctx = ctx

	queryHelper := baseapp.NewQueryServerTestHelper(suite.Ctx, suite.App.InterfaceRegistry())
	distrtypes.RegisterQueryServer(queryHelper, distrquery.NewQueryServer(
		tApp.GetDistrKeeper(),
		tApp.GetCommunityKeeper(),
	))
	suite.queryClient = distrtypes.NewQueryClient(queryHelper)
	suite.communityPoolAddress = tApp.GetAccountKeeper().GetModuleAddress(communitytypes.ModuleAccountName)
}

func TestGRPQueryTestSuite(t *testing.T) {
	suite.Run(t, new(queryTestSuite))
}

func (suite queryTestSuite) FundCommunityPool(amt sdk.Coins) {
	err := suite.App.FundModuleAccount(suite.Ctx, communitytypes.ModuleAccountName, amt)
	suite.NoError(err)
}

func (suite queryTestSuite) CheckCommunityPoolBalance(expected sdk.Coins, result sdk.DecCoins) {
	actual := suite.App.GetBankKeeper().GetAllBalances(suite.Ctx, suite.communityPoolAddress)
	// check that account was properly funded
	suite.Equal(expected, actual, "unexpected community pool balance")

	// transform the expected values to DecCoins to compare with result
	decCoins := sdk.NewDecCoinsFromCoins(expected...)
	suite.True(decCoins.IsEqual(result), "unexpected community pool query response")
}

func (suite *queryTestSuite) Test_CommunityPoolOverride() {
	singleDenom := sdk.NewCoins(sdk.NewInt64Coin("ukava", 1e10))
	multiDenom := singleDenom.Add(sdk.NewInt64Coin("other-denom", 1e9))

	testCases := []struct {
		name  string
		funds sdk.Coins
	}{
		{"single denom", singleDenom},
		{"multiple denoms", multiDenom},
		{"no balance", sdk.NewCoins()},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("queries the correct balance - %s", tc.name), func() {
			suite.SetupTest()

			// init community pool funds
			if !tc.funds.IsZero() {
				suite.FundCommunityPool(tc.funds)
			}

			// query the overridden endpoint
			balance, err := suite.queryClient.CommunityPool(
				context.Background(),
				&distrtypes.QueryCommunityPoolRequest{},
			)
			suite.NoError(err)
			suite.CheckCommunityPoolBalance(tc.funds, balance.Pool)
		})
	}
}
