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

// modified from sdk test of distribution querier
func (suite *queryTestSuite) Test_OGDistributionQueries() {
	suite.SetupTest()

	ctx, distrKeeper, stakingKeeper := suite.Ctx, suite.App.GetDistrKeeper(), suite.App.GetStakingKeeper()

	// setup a validator
	addr := app.RandomAddress()
	valAddr := sdk.ValAddress(addr)
	bondDenom := stakingKeeper.BondDenom(ctx)

	// test param queries
	params := distrtypes.Params{
		CommunityTax:        sdk.ZeroDec(),
		BaseProposerReward:  sdk.NewDecWithPrec(2, 1),
		BonusProposerReward: sdk.NewDecWithPrec(1, 1),
		WithdrawAddrEnabled: true,
	}
	distrKeeper.SetParams(ctx, params)
	r1, err := suite.queryClient.Params(context.Background(), &distrtypes.QueryParamsRequest{})
	suite.NoError(err)
	suite.Equal(params, r1.Params)

	// test outstanding rewards query
	outstandingRewards := sdk.NewDecCoins(sdk.NewInt64DecCoin(bondDenom, 100), sdk.NewInt64DecCoin("other", 10))
	distrKeeper.SetValidatorOutstandingRewards(ctx, valAddr, distrtypes.ValidatorOutstandingRewards{Rewards: outstandingRewards})
	r2, err := suite.queryClient.ValidatorOutstandingRewards(context.Background(), &distrtypes.QueryValidatorOutstandingRewardsRequest{ValidatorAddress: valAddr.String()})
	suite.NoError(err)
	suite.Equal(outstandingRewards, r2.Rewards.Rewards)

	// test validator commission query
	commission := sdk.DecCoins{{Denom: "token1", Amount: sdk.NewDec(4)}, {Denom: "token2", Amount: sdk.NewDec(2)}}
	distrKeeper.SetValidatorAccumulatedCommission(ctx, valAddr, distrtypes.ValidatorAccumulatedCommission{Commission: commission})
	r3, err := suite.queryClient.ValidatorCommission(context.Background(), &distrtypes.QueryValidatorCommissionRequest{ValidatorAddress: valAddr.String()})
	suite.NoError(err)
	suite.Equal(commission, r3.Commission.Commission)

	// test delegator's total rewards query
	r4, err := suite.queryClient.DelegationTotalRewards(context.Background(), &distrtypes.QueryDelegationTotalRewardsRequest{DelegatorAddress: addr.String()})
	suite.NoError(err)
	suite.Equal(&distrtypes.QueryDelegationTotalRewardsResponse{}, r4)

	// test validator slashes query with height range
	slashOne := distrtypes.NewValidatorSlashEvent(3, sdk.NewDecWithPrec(5, 1))
	slashTwo := distrtypes.NewValidatorSlashEvent(7, sdk.NewDecWithPrec(6, 1))
	distrKeeper.SetValidatorSlashEvent(ctx, valAddr, 3, 0, slashOne)
	distrKeeper.SetValidatorSlashEvent(ctx, valAddr, 7, 0, slashTwo)
	slashes := suite.getQueriedValidatorSlashes(valAddr, 0, 2)
	suite.Equal(0, len(slashes))
	slashes = suite.getQueriedValidatorSlashes(valAddr, 0, 5)
	suite.Equal([]distrtypes.ValidatorSlashEvent{slashOne}, slashes)
	slashes = suite.getQueriedValidatorSlashes(valAddr, 0, 10)
	suite.Equal([]distrtypes.ValidatorSlashEvent{slashOne, slashTwo}, slashes)

	// non-zero delegator reward queries are not tested here.
}

func (suite *queryTestSuite) getQueriedValidatorSlashes(validatorAddr sdk.ValAddress, startHeight uint64, endHeight uint64) (slashes []distrtypes.ValidatorSlashEvent) {
	result, err := suite.queryClient.ValidatorSlashes(
		context.Background(),
		&distrtypes.QueryValidatorSlashesRequest{
			ValidatorAddress: validatorAddr.String(),
			StartingHeight:   startHeight,
			EndingHeight:     endHeight,
		},
	)
	suite.NoError(err)
	return result.Slashes
}
