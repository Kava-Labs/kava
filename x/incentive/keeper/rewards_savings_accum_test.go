package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	savingskeeper "github.com/kava-labs/kava/x/savings/keeper"
	savingstypes "github.com/kava-labs/kava/x/savings/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/incentive/keeper"
	"github.com/kava-labs/kava/x/incentive/testutil"
	"github.com/kava-labs/kava/x/incentive/types"
)

// Test suite used for all keeper tests
type SavingsRewardsTestSuite struct {
	suite.Suite

	keeper        keeper.Keeper
	savingsKeeper savingskeeper.Keeper

	app app.TestApp
	ctx sdk.Context

	genesisTime time.Time
	addrs       []sdk.AccAddress
}

// SetupTest is run automatically before each suite test
func (suite *SavingsRewardsTestSuite) SetupTest() {
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)

	_, allAddrs := app.GeneratePrivKeyAddressPairs(10)
	suite.addrs = allAddrs[:5]
	suite.genesisTime = time.Date(2020, 12, 15, 14, 0, 0, 0, time.UTC)
}

func (suite *SavingsRewardsTestSuite) SetupApp() {
	suite.app = app.NewTestApp()

	suite.keeper = suite.app.GetIncentiveKeeper()
	suite.savingsKeeper = suite.app.GetSavingsKeeper()

	suite.ctx = suite.app.NewContext(true, tmproto.Header{Height: 1, Time: suite.genesisTime})
}

func (suite *SavingsRewardsTestSuite) SetupWithGenState(authBuilder *app.AuthBankGenesisBuilder, incentBuilder testutil.IncentiveGenesisBuilder,
	savingsGenesis savingstypes.GenesisState,
) {
	suite.SetupApp()

	suite.app.InitializeFromGenesisStatesWithTime(
		suite.genesisTime,
		authBuilder.BuildMarshalled(suite.app.AppCodec()),
		app.GenesisState{savingstypes.ModuleName: suite.app.AppCodec().MustMarshalJSON(&savingsGenesis)},
		incentBuilder.BuildMarshalled(suite.app.AppCodec()),
	)
}

func (suite *SavingsRewardsTestSuite) TestAccumulateSavingsRewards() {
	type args struct {
		deposit               sdk.Coin
		rewardsPerSecond      sdk.Coins
		timeElapsed           int
		expectedRewardIndexes types.RewardIndexes
	}
	type test struct {
		name string
		args args
	}
	testCases := []test{
		{
			"7 seconds",
			args{
				deposit:          c("ukava", 1_000_000),
				rewardsPerSecond: cs(c("hard", 122354)),
				timeElapsed:      7,
				expectedRewardIndexes: types.RewardIndexes{
					types.NewRewardIndex("hard", d("0.856478000000000000")),
				},
			},
		},
		{
			"1 day",
			args{
				deposit:          c("ukava", 1_000_000),
				rewardsPerSecond: cs(c("hard", 122354)),
				timeElapsed:      86400,
				expectedRewardIndexes: types.RewardIndexes{
					types.NewRewardIndex("hard", d("10571.385600000000000000")),
				},
			},
		},
		{
			"0 seconds",
			args{
				deposit:          c("ukava", 1_000_000),
				rewardsPerSecond: cs(c("hard", 122354)),
				timeElapsed:      0,
				expectedRewardIndexes: types.RewardIndexes{
					types.NewRewardIndex("hard", d("0.0")),
				},
			},
		},
		{
			"multiple reward coins",
			args{
				deposit:          c("ukava", 1_000_000),
				rewardsPerSecond: cs(c("hard", 122354), c("bnb", 567889)),
				timeElapsed:      7,
				expectedRewardIndexes: types.RewardIndexes{
					types.NewRewardIndex("bnb", d("3.97522300000000000")),
					types.NewRewardIndex("hard", d("0.856478000000000000")),
				},
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			params := savingstypes.NewParams(
				[]string{"ukava"},
			)
			deposits := savingstypes.Deposits{
				savingstypes.NewDeposit(
					suite.addrs[0],
					sdk.NewCoins(tc.args.deposit),
				),
			}
			savingsGenesis := savingstypes.NewGenesisState(params, deposits)

			authBuilder := app.NewAuthBankGenesisBuilder().
				WithSimpleAccount(suite.addrs[0], cs(c("ukava", 1e9))).
				WithSimpleModuleAccount(savingstypes.ModuleName, sdk.NewCoins(tc.args.deposit))

			incentBuilder := testutil.NewIncentiveGenesisBuilder().
				WithGenesisTime(suite.genesisTime).
				WithSimpleSavingsRewardPeriod(tc.args.deposit.Denom, tc.args.rewardsPerSecond)

			suite.SetupWithGenState(authBuilder, incentBuilder, savingsGenesis)

			// Set up chain context at future time
			runAtTime := suite.ctx.BlockTime().Add(time.Duration(int(time.Second) * tc.args.timeElapsed))
			runCtx := suite.ctx.WithBlockTime(runAtTime)

			rewardPeriods, found := suite.keeper.GetSavingsRewardPeriods(runCtx, tc.args.deposit.Denom)
			suite.Require().True(found)
			suite.keeper.AccumulateSavingsRewards(runCtx, rewardPeriods)

			rewardIndexes, _ := suite.keeper.GetSavingsRewardIndexes(runCtx, tc.args.deposit.Denom)
			suite.Require().Equal(tc.args.expectedRewardIndexes, rewardIndexes)
		})
	}
}

func TestSavingsRewardsTestSuite(t *testing.T) {
	suite.Run(t, new(SavingsRewardsTestSuite))
}
