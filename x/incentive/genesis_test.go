package incentive_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/hard"
	"github.com/kava-labs/kava/x/incentive"
	"github.com/kava-labs/kava/x/kavadist"
)

const (
	oneYear time.Duration = 365 * 24 * time.Hour
)

type GenesisTestSuite struct {
	suite.Suite

	ctx    sdk.Context
	app    app.TestApp
	keeper incentive.Keeper
	addrs  []sdk.AccAddress

	genesisTime time.Time
}

func (suite *GenesisTestSuite) SetupTest() {
	tApp := app.NewTestApp()
	keeper := tApp.GetIncentiveKeeper()
	suite.genesisTime = time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)

	_, addrs := app.GeneratePrivKeyAddressPairs(3)
	coins := []sdk.Coins{}
	for j := 0; j < 3; j++ {
		coins = append(coins, cs(c("bnb", 10_000_000_000), c("ukava", 10_000_000_000)))
	}
	authGS := app.NewAuthGenState(addrs, coins)

	loanToValue, _ := sdk.NewDecFromStr("0.6")
	borrowLimit := sdk.NewDec(1000000000000000)
	hardGS := hard.NewGenesisState(
		hard.NewParams(
			hard.MoneyMarkets{
				hard.NewMoneyMarket("ukava", hard.NewBorrowLimit(false, borrowLimit, loanToValue), "kava:usd", sdk.NewInt(1000000), hard.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
				hard.NewMoneyMarket("bnb", hard.NewBorrowLimit(false, borrowLimit, loanToValue), "bnb:usd", sdk.NewInt(1000000), hard.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
			},
			sdk.NewDec(10),
		),
		hard.DefaultAccumulationTimes,
		hard.DefaultDeposits,
		hard.DefaultBorrows,
		hard.DefaultTotalSupplied,
		hard.DefaultTotalBorrowed,
		hard.DefaultTotalReserves,
	)
	incentiveGS := incentive.NewGenesisState(
		incentive.NewParams(
			incentive.RewardPeriods{incentive.NewRewardPeriod(true, "bnb-a", suite.genesisTime.Add(-1*oneYear), suite.genesisTime.Add(oneYear), c("ukava", 122354))},
			incentive.MultiRewardPeriods{incentive.NewMultiRewardPeriod(true, "bnb", suite.genesisTime.Add(-1*oneYear), suite.genesisTime.Add(oneYear), cs(c("hard", 122354)))},
			incentive.MultiRewardPeriods{incentive.NewMultiRewardPeriod(true, "bnb", suite.genesisTime.Add(-1*oneYear), suite.genesisTime.Add(oneYear), cs(c("hard", 122354)))},
			incentive.RewardPeriods{incentive.NewRewardPeriod(true, "ukava", suite.genesisTime.Add(-1*oneYear), suite.genesisTime.Add(oneYear), c("hard", 122354))},
			incentive.Multipliers{incentive.NewMultiplier(incentive.Small, 1, d("0.25")), incentive.NewMultiplier(incentive.Large, 12, d("1.0"))},
			suite.genesisTime.Add(5*oneYear),
		),
		incentive.DefaultGenesisAccumulationTimes,
		incentive.DefaultGenesisAccumulationTimes,
		incentive.DefaultGenesisAccumulationTimes,
		incentive.DefaultGenesisAccumulationTimes,
		incentive.DefaultGenesisRewardIndexes,
		incentive.DefaultGenesisRewardIndexes,
		incentive.DefaultGenesisRewardIndexes,
		incentive.DefaultGenesisRewardIndexes,
		incentive.DefaultUSDXClaims,
		incentive.DefaultHardClaims,
	)
	tApp.InitializeFromGenesisStatesWithTime(
		suite.genesisTime,
		authGS,
		app.GenesisState{incentive.ModuleName: incentive.ModuleCdc.MustMarshalJSON(incentiveGS)},
		app.GenesisState{hard.ModuleName: hard.ModuleCdc.MustMarshalJSON(hardGS)},
		NewCDPGenStateMulti(),
		NewPricefeedGenStateMulti(),
	)

	ctx := tApp.NewContext(true, abci.Header{Height: 1, Time: suite.genesisTime})

	// TODO add to auth gen state to tidy up test
	err := tApp.GetSupplyKeeper().MintCoins(
		ctx,
		kavadist.KavaDistMacc,
		cs(c("hard", 1_000_000_000_000_000), c("ukava", 1_000_000_000_000_000)),
	)
	suite.Require().NoError(err)

	suite.addrs = addrs
	suite.keeper = keeper
	suite.app = tApp
	suite.ctx = ctx
}

// Test to cover an bug where paid out claims would zero out rewards incorrectly, creating an invalid coins object.
// The invalid reward coins would fail the genesis state validation
func (suite *GenesisTestSuite) TestPaidOutClaimsPassValidateGenesis() {
	hardHandler := hard.NewHandler(suite.app.GetHardKeeper())
	_, err := hardHandler(suite.ctx, hard.NewMsgDeposit(suite.addrs[0], cs(c("bnb", 100_000_000))))
	suite.Require().NoError(err)

	suite.app.EndBlocker(suite.ctx, abci.RequestEndBlock{})
	suite.ctx = suite.ctx.WithBlockTime(suite.genesisTime.Add(1 * 10 * time.Second))
	suite.app.BeginBlocker(suite.ctx, abci.RequestBeginBlock{})

	suite.app.EndBlocker(suite.ctx, abci.RequestEndBlock{})
	suite.ctx = suite.ctx.WithBlockTime(suite.genesisTime.Add(2 * 10 * time.Second))
	suite.app.BeginBlocker(suite.ctx, abci.RequestBeginBlock{})

	_, err = hardHandler(suite.ctx, hard.NewMsgWithdraw(suite.addrs[0], cs(c("bnb", 100_000_000))))
	suite.Require().NoError(err)

	incentiveHandler := incentive.NewHandler(suite.keeper)
	_, err = incentiveHandler(suite.ctx, incentive.NewMsgClaimHardReward(suite.addrs[0], string(incentive.Large)))
	suite.Require().NoError(err)

	genState := incentive.ExportGenesis(suite.ctx, suite.keeper)
	suite.Require().NoError(
		genState.Validate(),
	)
}

func (suite *GenesisTestSuite) TestExportedGenesisMatchesImported() {
	genesisTime := time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)
	genesisState := incentive.NewGenesisState(
		incentive.NewParams(
			incentive.RewardPeriods{incentive.NewRewardPeriod(true, "bnb-a", genesisTime.Add(-1*oneYear), genesisTime.Add(oneYear), c("ukava", 122354))},
			incentive.MultiRewardPeriods{incentive.NewMultiRewardPeriod(true, "bnb", genesisTime.Add(-1*oneYear), genesisTime.Add(oneYear), cs(c("hard", 122354)))},
			incentive.MultiRewardPeriods{incentive.NewMultiRewardPeriod(true, "bnb", genesisTime.Add(-1*oneYear), genesisTime.Add(oneYear), cs(c("hard", 122354)))},
			incentive.RewardPeriods{incentive.NewRewardPeriod(true, "ukava", genesisTime.Add(-1*oneYear), genesisTime.Add(oneYear), c("hard", 122354))},
			incentive.Multipliers{incentive.NewMultiplier(incentive.Small, 1, d("0.25")), incentive.NewMultiplier(incentive.Large, 12, d("1.0"))},
			genesisTime.Add(5*oneYear),
		),
		incentive.GenesisAccumulationTimes{
			incentive.NewGenesisAccumulationTime("bnb-a", genesisTime),
		},
		incentive.GenesisAccumulationTimes{
			incentive.NewGenesisAccumulationTime("bnb", genesisTime.Add(-1*time.Hour)),
		},
		incentive.GenesisAccumulationTimes{
			incentive.NewGenesisAccumulationTime("bnb", genesisTime.Add(-2*time.Hour)),
		},
		incentive.GenesisAccumulationTimes{
			incentive.NewGenesisAccumulationTime("ukava", genesisTime.Add(-3*time.Hour)),
		},
		incentive.GenesisRewardIndexesSlice{
			incentive.NewGenesisRewardIndexes("bnb-a", incentive.RewardIndexes{{CollateralType: "ukava", RewardFactor: d("0.3")}}),
		},
		incentive.GenesisRewardIndexesSlice{
			incentive.NewGenesisRewardIndexes("bnb", incentive.RewardIndexes{{CollateralType: "hard", RewardFactor: d("0.1")}}),
		},
		incentive.GenesisRewardIndexesSlice{
			incentive.NewGenesisRewardIndexes("bnb", incentive.RewardIndexes{{CollateralType: "hard", RewardFactor: d("0.05")}}),
		},
		incentive.GenesisRewardIndexesSlice{
			incentive.NewGenesisRewardIndexes("ukava", incentive.RewardIndexes{{CollateralType: "hard", RewardFactor: d("0.2")}}),
		},
		incentive.USDXMintingClaims{
			incentive.NewUSDXMintingClaim(
				suite.addrs[0],
				c("ukava", 1e9),
				incentive.RewardIndexes{{CollateralType: "bnb-a", RewardFactor: d("0.3")}},
			),
			incentive.NewUSDXMintingClaim(
				suite.addrs[1],
				c("ukava", 1),
				incentive.RewardIndexes{{CollateralType: "bnb-a", RewardFactor: d("0.001")}},
			),
		},
		incentive.HardLiquidityProviderClaims{
			incentive.NewHardLiquidityProviderClaim(
				suite.addrs[0],
				cs(c("ukava", 1e9), c("hard", 1e9)),
				incentive.MultiRewardIndexes{{CollateralType: "bnb", RewardIndexes: incentive.RewardIndexes{{CollateralType: "hard", RewardFactor: d("0.01")}}}},
				incentive.MultiRewardIndexes{{CollateralType: "bnb", RewardIndexes: incentive.RewardIndexes{{CollateralType: "hard", RewardFactor: d("0.0")}}}},
				incentive.RewardIndexes{{CollateralType: "ukava", RewardFactor: d("0.2")}},
			),
			incentive.NewHardLiquidityProviderClaim(
				suite.addrs[1],
				cs(c("hard", 1)),
				incentive.MultiRewardIndexes{{CollateralType: "bnb", RewardIndexes: incentive.RewardIndexes{{CollateralType: "hard", RewardFactor: d("0.1")}}}},
				incentive.MultiRewardIndexes{{CollateralType: "bnb", RewardIndexes: incentive.RewardIndexes{{CollateralType: "hard", RewardFactor: d("0.0")}}}},
				incentive.RewardIndexes{{CollateralType: "ukava", RewardFactor: d("0.0")}},
			),
		},
	)

	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, abci.Header{Height: 1})

	// Incentive init genesis reads from the cdp keeper to check params are ok. So it needs to be initialized first.
	// Then the cdp keeper reads from pricefeed keeper to check its params are ok. So it also need initialization.
	tApp.InitializeFromGenesisStates(
		NewCDPGenStateMulti(),
		NewPricefeedGenStateMulti(),
	)

	incentive.InitGenesis(ctx, tApp.GetIncentiveKeeper(), tApp.GetSupplyKeeper(), tApp.GetCDPKeeper(), genesisState)

	exportedGenesisState := incentive.ExportGenesis(ctx, tApp.GetIncentiveKeeper())

	suite.Equal(genesisState, exportedGenesisState)
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}
