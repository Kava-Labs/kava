package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/kava-labs/kava/x/community/keeper"
	"github.com/kava-labs/kava/x/community/testutil"
	"github.com/kava-labs/kava/x/community/types"
	hardkeeper "github.com/kava-labs/kava/x/hard/keeper"
	hardtypes "github.com/kava-labs/kava/x/hard/types"
	pricefeedtypes "github.com/kava-labs/kava/x/pricefeed/types"
)

func ukava(amt int64) sdk.Coins {
	return sdk.NewCoins(sdk.NewInt64Coin("ukava", amt))
}
func usdx(amt int64) sdk.Coins {
	return sdk.NewCoins(sdk.NewInt64Coin("usdx", amt))
}
func otherdenom(amt int64) sdk.Coins {
	return sdk.NewCoins(sdk.NewInt64Coin("other-denom", amt))
}

type proposalTestSuite struct {
	testutil.Suite

	hardKeeper hardkeeper.Keeper
}

func TestProposalTestSuite(t *testing.T) {
	suite.Run(t, new(proposalTestSuite))
}

func (suite *proposalTestSuite) SetupTest() {
	suite.Suite.SetupTest()

	// setup money markets for kava & usdx
	suite.hardKeeper = suite.App.GetHardKeeper()
	suite.SetupMoneyMarket("ukava", "kava:usd")
	suite.SetupMoneyMarket("usdx", "usdx:usd")

	// give the community pool some funds
	// ukava
	err := suite.App.FundModuleAccount(suite.Ctx, types.ModuleAccountName, ukava(1e10))
	suite.NoError(err)

	// usdx
	err = suite.App.FundModuleAccount(suite.Ctx, types.ModuleAccountName, usdx(1e10))
	suite.NoError(err)

	// other-denom
	err = suite.App.FundModuleAccount(suite.Ctx, types.ModuleAccountName, otherdenom(1e10))
	suite.NoError(err)
}

func (suite *proposalTestSuite) SetupMoneyMarket(denom, spotMarketId string) {
	// add money market to hard
	suite.hardKeeper.SetMoneyMarket(
		suite.Ctx,
		denom,
		hardtypes.NewMoneyMarket(denom, hardtypes.NewBorrowLimit(false, sdk.NewDec(1e15), sdk.MustNewDecFromStr("0.6")), spotMarketId, sdk.NewInt(1e6), hardtypes.NewInterestRateModel(sdk.MustNewDecFromStr("0.05"), sdk.MustNewDecFromStr("2"), sdk.MustNewDecFromStr("0.8"), sdk.MustNewDecFromStr("10")), sdk.MustNewDecFromStr("0.05"), sdk.ZeroDec()),
	)

	// setup pricefeed
	pfk := suite.App.GetPriceFeedKeeper()
	pfk.SetParams(suite.Ctx, pricefeedtypes.NewParams(
		[]pricefeedtypes.Market{
			{MarketID: spotMarketId, BaseAsset: denom, QuoteAsset: "usd", Oracles: []sdk.AccAddress{}, Active: true},
		},
	))
	_, err := pfk.SetPrice(suite.Ctx, sdk.AccAddress{}, spotMarketId, sdk.OneDec(), time.Now().Add(1*time.Hour))
	suite.NoError(err)
}

func (suite *proposalTestSuite) TestCommunityLendDepositProposal() {
	testCases := []struct {
		name             string
		proposals        []*types.CommunityPoolLendDepositProposal
		expectedErr      string
		expectedDeposits []sdk.Coins
	}{
		{
			name: "valid - one proposal",
			proposals: []*types.CommunityPoolLendDepositProposal{
				{Amount: ukava(1e8)},
			},
			expectedErr:      "",
			expectedDeposits: []sdk.Coins{ukava(1e8)},
		},
		{
			name: "valid - multiple proposals, same denom",
			proposals: []*types.CommunityPoolLendDepositProposal{
				{Amount: ukava(1e8)},
				{Amount: ukava(1e9)},
			},
			expectedErr:      "",
			expectedDeposits: []sdk.Coins{ukava(1e8 + 1e9)},
		},
		{
			name: "valid - multiple proposals, different denoms",
			proposals: []*types.CommunityPoolLendDepositProposal{
				{Amount: ukava(1e8)},
				{Amount: usdx(1e8)},
			},
			expectedErr:      "",
			expectedDeposits: []sdk.Coins{ukava(1e8).Add(usdx(1e8)...)},
		},
		{
			name: "invalid - insufficient balance",
			proposals: []*types.CommunityPoolLendDepositProposal{
				{
					Description: "more coins than i have!",
					Amount:      ukava(1e11),
				},
			},
			expectedErr:      "insufficient funds",
			expectedDeposits: []sdk.Coins{},
		},
		{
			name: "invalid - invalid lend deposit (unsupported denom)",
			proposals: []*types.CommunityPoolLendDepositProposal{
				{Amount: otherdenom(1e9)},
			},
			expectedErr:      "invalid deposit denom",
			expectedDeposits: []sdk.Coins{},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			for _, proposal := range tc.proposals {
				err := keeper.HandleCommunityPoolLendDepositProposal(suite.Ctx, suite.Keeper, proposal)
				if tc.expectedErr == "" {
					suite.NoError(err)
				} else {
					suite.ErrorContains(err, tc.expectedErr)
				}
			}

			deposits := suite.hardKeeper.GetDepositsByUser(suite.Ctx, suite.MaccAddress)
			suite.Len(deposits, len(tc.expectedDeposits), "expected a deposit to lend")
			for _, amt := range tc.expectedDeposits {
				suite.Equal(amt, deposits[0].Amount, "expected amount to match")
			}
		})
	}
}
