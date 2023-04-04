package keeper_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	hardtypes "github.com/kava-labs/kava/x/hard/types"
	"github.com/kava-labs/kava/x/incentive/keeper"
	"github.com/kava-labs/kava/x/incentive/types"
	"github.com/stretchr/testify/suite"
	tmprototypes "github.com/tendermint/tendermint/proto/tendermint/types"
)

const (
	oneYear time.Duration = 365 * 24 * time.Hour
)

type grpcQueryTestSuite struct {
	suite.Suite

	tApp        app.TestApp
	ctx         sdk.Context
	keeper      keeper.Keeper
	queryClient types.QueryClient
	addrs       []sdk.AccAddress

	genesisTime  time.Time
	genesisState types.GenesisState
}

func (suite *grpcQueryTestSuite) SetupTest() {
	suite.tApp = app.NewTestApp()
	cdc := suite.tApp.AppCodec()

	_, addrs := app.GeneratePrivKeyAddressPairs(5)
	suite.genesisTime = time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)

	suite.addrs = addrs

	suite.ctx = suite.tApp.NewContext(true, tmprototypes.Header{}).
		WithBlockTime(time.Now().UTC())
	suite.keeper = suite.tApp.GetIncentiveKeeper()

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.tApp.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, keeper.NewQueryServerImpl(suite.keeper))

	suite.queryClient = types.NewQueryClient(queryHelper)

	loanToValue, _ := sdk.NewDecFromStr("0.6")
	borrowLimit := sdk.NewDec(1000000000000000)
	hardGS := hardtypes.NewGenesisState(
		hardtypes.NewParams(
			hardtypes.MoneyMarkets{
				hardtypes.NewMoneyMarket("ukava", hardtypes.NewBorrowLimit(false, borrowLimit, loanToValue), "kava:usd", sdk.NewInt(1000000), hardtypes.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
				hardtypes.NewMoneyMarket("bnb", hardtypes.NewBorrowLimit(false, borrowLimit, loanToValue), "bnb:usd", sdk.NewInt(1000000), hardtypes.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
			},
			sdk.NewDec(10),
		),
		hardtypes.DefaultAccumulationTimes,
		hardtypes.DefaultDeposits,
		hardtypes.DefaultBorrows,
		hardtypes.DefaultTotalSupplied,
		hardtypes.DefaultTotalBorrowed,
		hardtypes.DefaultTotalReserves,
	)

	suite.genesisState = types.NewGenesisState(
		types.NewParams(
			types.RewardPeriods{types.NewRewardPeriod(true, "bnb-a", suite.genesisTime.Add(-1*oneYear), suite.genesisTime.Add(oneYear), c("ukava", 122354))},
			types.MultiRewardPeriods{types.NewMultiRewardPeriod(true, "bnb", suite.genesisTime.Add(-1*oneYear), suite.genesisTime.Add(oneYear), cs(c("hard", 122354)))},
			types.MultiRewardPeriods{types.NewMultiRewardPeriod(true, "bnb", suite.genesisTime.Add(-1*oneYear), suite.genesisTime.Add(oneYear), cs(c("hard", 122354)))},
			types.MultiRewardPeriods{types.NewMultiRewardPeriod(true, "ukava", suite.genesisTime.Add(-1*oneYear), suite.genesisTime.Add(oneYear), cs(c("hard", 122354)))},
			types.MultiRewardPeriods{types.NewMultiRewardPeriod(true, "btcb/usdx", suite.genesisTime.Add(-1*oneYear), suite.genesisTime.Add(oneYear), cs(c("swp", 122354)))},
			types.MultiRewardPeriods{types.NewMultiRewardPeriod(true, "ukava", suite.genesisTime.Add(-1*oneYear), suite.genesisTime.Add(oneYear), cs(c("hard", 122354)))},
			types.MultiRewardPeriods{types.NewMultiRewardPeriod(true, "ukava", suite.genesisTime.Add(-1*oneYear), suite.genesisTime.Add(oneYear), cs(c("hard", 122354)))},
			types.MultipliersPerDenoms{
				{
					Denom: "ukava",
					Multipliers: types.Multipliers{
						types.NewMultiplier("large", 12, d("1.0")),
					},
				},
				{
					Denom: "hard",
					Multipliers: types.Multipliers{
						types.NewMultiplier("small", 1, d("0.25")),
						types.NewMultiplier("large", 12, d("1.0")),
					},
				},
				{
					Denom: "swp",
					Multipliers: types.Multipliers{
						types.NewMultiplier("small", 1, d("0.25")),
						types.NewMultiplier("medium", 6, d("0.8")),
					},
				},
			},
			suite.genesisTime.Add(5*oneYear),
		),
		types.NewGenesisRewardState(
			types.AccumulationTimes{
				types.NewAccumulationTime("bnb-a", suite.genesisTime),
			},
			types.MultiRewardIndexes{
				types.NewMultiRewardIndex("bnb-a", types.RewardIndexes{{CollateralType: "ukava", RewardFactor: d("0.3")}}),
			},
		),
		types.NewGenesisRewardState(
			types.AccumulationTimes{
				types.NewAccumulationTime("bnb", suite.genesisTime.Add(-1*time.Hour)),
			},
			types.MultiRewardIndexes{
				types.NewMultiRewardIndex("bnb", types.RewardIndexes{{CollateralType: "hard", RewardFactor: d("0.1")}}),
			},
		),
		types.NewGenesisRewardState(
			types.AccumulationTimes{
				types.NewAccumulationTime("bnb", suite.genesisTime.Add(-2*time.Hour)),
			},
			types.MultiRewardIndexes{
				types.NewMultiRewardIndex("bnb", types.RewardIndexes{{CollateralType: "hard", RewardFactor: d("0.05")}}),
			},
		),
		types.NewGenesisRewardState(
			types.AccumulationTimes{
				types.NewAccumulationTime("ukava", suite.genesisTime.Add(-3*time.Hour)),
			},
			types.MultiRewardIndexes{
				types.NewMultiRewardIndex("ukava", types.RewardIndexes{{CollateralType: "hard", RewardFactor: d("0.2")}}),
			},
		),
		types.NewGenesisRewardState(
			types.AccumulationTimes{
				types.NewAccumulationTime("bctb/usdx", suite.genesisTime.Add(-4*time.Hour)),
			},
			types.MultiRewardIndexes{
				types.NewMultiRewardIndex("btcb/usdx", types.RewardIndexes{{CollateralType: "swap", RewardFactor: d("0.001")}}),
			},
		),
		types.NewGenesisRewardState(
			types.AccumulationTimes{
				types.NewAccumulationTime("ukava", suite.genesisTime.Add(-3*time.Hour)),
			},
			types.MultiRewardIndexes{
				types.NewMultiRewardIndex("ukava", types.RewardIndexes{{CollateralType: "ukava", RewardFactor: d("0.2")}}),
			},
		),
		types.NewGenesisRewardState(
			types.AccumulationTimes{
				types.NewAccumulationTime("usdx", suite.genesisTime.Add(-3*time.Hour)),
			},
			types.MultiRewardIndexes{
				types.NewMultiRewardIndex("usdx", types.RewardIndexes{{CollateralType: "usdx", RewardFactor: d("0.2")}}),
			},
		),
		types.USDXMintingClaims{
			types.NewUSDXMintingClaim(
				suite.addrs[0],
				c("ukava", 1e9),
				types.RewardIndexes{{CollateralType: "bnb-a", RewardFactor: d("0.3")}},
			),
			types.NewUSDXMintingClaim(
				suite.addrs[1],
				c("ukava", 1),
				types.RewardIndexes{{CollateralType: "bnb-a", RewardFactor: d("0.001")}},
			),
		},
		types.HardLiquidityProviderClaims{
			types.NewHardLiquidityProviderClaim(
				suite.addrs[0],
				cs(c("ukava", 1e9), c("hard", 1e9)),
				types.MultiRewardIndexes{{CollateralType: "bnb", RewardIndexes: types.RewardIndexes{{CollateralType: "hard", RewardFactor: d("0.01")}}}},
				types.MultiRewardIndexes{{CollateralType: "bnb", RewardIndexes: types.RewardIndexes{{CollateralType: "hard", RewardFactor: d("0.0")}}}},
			),
			types.NewHardLiquidityProviderClaim(
				suite.addrs[1],
				cs(c("hard", 1)),
				types.MultiRewardIndexes{{CollateralType: "bnb", RewardIndexes: types.RewardIndexes{{CollateralType: "hard", RewardFactor: d("0.1")}}}},
				types.MultiRewardIndexes{{CollateralType: "bnb", RewardIndexes: types.RewardIndexes{{CollateralType: "hard", RewardFactor: d("0.0")}}}},
			),
		},
		types.DelegatorClaims{
			types.NewDelegatorClaim(
				suite.addrs[2],
				cs(c("hard", 5)),
				types.MultiRewardIndexes{{CollateralType: "ukava", RewardIndexes: types.RewardIndexes{{CollateralType: "hard", RewardFactor: d("0.2")}}}},
			),
		},
		types.SwapClaims{
			types.NewSwapClaim(
				suite.addrs[3],
				nil,
				types.MultiRewardIndexes{{CollateralType: "btcb/usdx", RewardIndexes: types.RewardIndexes{{CollateralType: "swap", RewardFactor: d("0.0")}}}},
			),
		},
		types.SavingsClaims{
			types.NewSavingsClaim(
				suite.addrs[3],
				nil,
				types.MultiRewardIndexes{{CollateralType: "ukava", RewardIndexes: types.RewardIndexes{{CollateralType: "ukava", RewardFactor: d("0.0")}}}},
			),
		},
		types.EarnClaims{
			types.NewEarnClaim(
				suite.addrs[3],
				nil,
				types.MultiRewardIndexes{{CollateralType: "usdx", RewardIndexes: types.RewardIndexes{{CollateralType: "usdx", RewardFactor: d("0.0")}}}},
			),
		},
	)

	err := suite.genesisState.Validate()
	suite.Require().NoError(err)

	suite.tApp = suite.tApp.InitializeFromGenesisStatesWithTime(
		suite.genesisTime,
		app.GenesisState{types.ModuleName: cdc.MustMarshalJSON(&suite.genesisState)},
		app.GenesisState{hardtypes.ModuleName: cdc.MustMarshalJSON(&hardGS)},
		NewCDPGenStateMulti(cdc),
		NewPricefeedGenStateMultiFromTime(cdc, suite.genesisTime),
	)

	suite.tApp.DeleteGenesisValidator(suite.T(), suite.ctx)
	claims := suite.keeper.GetAllDelegatorClaims(suite.ctx)
	for _, claim := range claims {
		// Delete the InitGenesis validator's claim
		if !claim.Owner.Equals(suite.addrs[2]) {
			suite.keeper.DeleteDelegatorClaim(suite.ctx, claim.Owner)
		}
	}
}

func (suite *grpcQueryTestSuite) TestGrpcQueryParams() {
	res, err := suite.queryClient.Params(sdk.WrapSDKContext(suite.ctx), &types.QueryParamsRequest{})
	suite.Require().NoError(err)

	expected := suite.keeper.GetParams(suite.ctx)

	suite.Equal(expected, res.Params, "params should equal default params")
}

func (suite *grpcQueryTestSuite) TestGrpcQueryRewards() {
	res, err := suite.queryClient.Rewards(sdk.WrapSDKContext(suite.ctx), &types.QueryRewardsRequest{
		Unsynchronized: true,
	})
	suite.Require().NoError(err)

	suite.Equal(suite.genesisState.USDXMintingClaims, res.USDXMintingClaims)
	suite.Equal(suite.genesisState.HardLiquidityProviderClaims, res.HardLiquidityProviderClaims)
	suite.Equal(suite.genesisState.DelegatorClaims, res.DelegatorClaims)
	suite.Equal(suite.genesisState.SwapClaims, res.SwapClaims)
	suite.Equal(suite.genesisState.SavingsClaims, res.SavingsClaims)
	suite.Equal(suite.genesisState.EarnClaims, res.EarnClaims)
}

func (suite *grpcQueryTestSuite) TestGrpcQueryRewards_Owner() {
	res, err := suite.queryClient.Rewards(sdk.WrapSDKContext(suite.ctx), &types.QueryRewardsRequest{
		Owner: suite.addrs[0].String(),
	})
	suite.Require().NoError(err)

	suite.Len(res.USDXMintingClaims, 1)
	suite.Len(res.HardLiquidityProviderClaims, 1)

	suite.Equal(suite.genesisState.USDXMintingClaims[0], res.USDXMintingClaims[0])
	suite.Equal(suite.genesisState.HardLiquidityProviderClaims[0], res.HardLiquidityProviderClaims[0])

	// No other claims - owner has none
	suite.Empty(res.DelegatorClaims)
	suite.Empty(res.SwapClaims)
	suite.Empty(res.SavingsClaims)
	suite.Empty(res.EarnClaims)
}

func (suite *grpcQueryTestSuite) TestGrpcQueryRewards_RewardType() {
	res, err := suite.queryClient.Rewards(sdk.WrapSDKContext(suite.ctx), &types.QueryRewardsRequest{
		RewardType:     keeper.RewardTypeHard,
		Unsynchronized: true,
	})
	suite.Require().NoError(err)

	suite.Equal(suite.genesisState.HardLiquidityProviderClaims, res.HardLiquidityProviderClaims)

	// No other reward types when specifying rewardType
	suite.Empty(res.USDXMintingClaims)
	suite.Empty(res.DelegatorClaims)
	suite.Empty(res.SwapClaims)
	suite.Empty(res.SavingsClaims)
	suite.Empty(res.EarnClaims)
}

func (suite *grpcQueryTestSuite) TestGrpcQueryRewards_RewardType_and_Owner() {
	res, err := suite.queryClient.Rewards(sdk.WrapSDKContext(suite.ctx), &types.QueryRewardsRequest{
		Owner:      suite.addrs[0].String(),
		RewardType: keeper.RewardTypeHard,
	})
	suite.Require().NoError(err)

	suite.Len(res.HardLiquidityProviderClaims, 1)
	suite.Equal(suite.genesisState.HardLiquidityProviderClaims[0], res.HardLiquidityProviderClaims[0])

	suite.Empty(res.USDXMintingClaims)
	suite.Empty(res.DelegatorClaims)
	suite.Empty(res.SwapClaims)
	suite.Empty(res.SavingsClaims)
	suite.Empty(res.EarnClaims)
}

func (suite *grpcQueryTestSuite) TestGrpcQueryRewardFactors() {
	res, err := suite.queryClient.RewardFactors(sdk.WrapSDKContext(suite.ctx), &types.QueryRewardFactorsRequest{})
	suite.Require().NoError(err)

	suite.NotEmpty(res.UsdxMintingRewardFactors)
	suite.NotEmpty(res.HardSupplyRewardFactors)
	suite.NotEmpty(res.HardBorrowRewardFactors)
	suite.NotEmpty(res.DelegatorRewardFactors)
	suite.NotEmpty(res.SwapRewardFactors)
	suite.NotEmpty(res.SavingsRewardFactors)
	suite.NotEmpty(res.EarnRewardFactors)
}

func TestGrpcQueryTestSuite(t *testing.T) {
	suite.Run(t, new(grpcQueryTestSuite))
}
