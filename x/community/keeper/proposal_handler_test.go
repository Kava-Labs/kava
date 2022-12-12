package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/community/keeper"
	"github.com/kava-labs/kava/x/community/testutil"
	"github.com/kava-labs/kava/x/community/types"
	hardkeeper "github.com/kava-labs/kava/x/hard/keeper"
	hardtypes "github.com/kava-labs/kava/x/hard/types"
	pricefeedtypes "github.com/kava-labs/kava/x/pricefeed/types"
)

const chainID = "kavatest_2221-1"

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
	suite.Suite

	App         app.TestApp
	Ctx         sdk.Context
	Keeper      keeper.Keeper
	MaccAddress sdk.AccAddress

	hardKeeper hardkeeper.Keeper
}

func TestProposalTestSuite(t *testing.T) {
	suite.Run(t, new(proposalTestSuite))
}

func (suite *proposalTestSuite) SetupTest() {
	app.SetSDKConfig()

	genTime := tmtime.Now()

	hardGS, pricefeedGS := testutil.NewLendGenesisBuilder().
		WithMarket("ukava", "kava:usd", sdk.OneDec()).
		WithMarket("usdx", "usdx:usd", sdk.OneDec()).
		Build()

	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, tmproto.Header{
		Height:  1,
		Time:    genTime,
		ChainID: chainID,
	})

	tApp.InitializeFromGenesisStatesWithTimeAndChainID(
		genTime, chainID,
		app.GenesisState{hardtypes.ModuleName: tApp.AppCodec().MustMarshalJSON(&hardGS)},
		app.GenesisState{pricefeedtypes.ModuleName: tApp.AppCodec().MustMarshalJSON(&pricefeedGS)},
	)

	suite.App = tApp
	suite.Ctx = ctx
	suite.Keeper = tApp.GetCommunityKeeper()
	suite.MaccAddress = tApp.GetAccountKeeper().GetModuleAddress(types.ModuleAccountName)
	suite.hardKeeper = suite.App.GetHardKeeper()

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

func (suite *proposalTestSuite) TestCommunityLendDepositProposal() {
	testCases := []struct {
		name             string
		proposals        []*types.CommunityPoolLendDepositProposal
		expectedErr      string
		expectedDeposits []sdk.Coins
	}{
		{
			name: "valid - one proposal, one denom",
			proposals: []*types.CommunityPoolLendDepositProposal{
				{Amount: ukava(1e8)},
			},
			expectedErr:      "",
			expectedDeposits: []sdk.Coins{ukava(1e8)},
		},
		{
			name: "valid - one proposal, multiple denoms",
			proposals: []*types.CommunityPoolLendDepositProposal{
				{Amount: ukava(1e8).Add(usdx(1e8)...)},
			},
			expectedErr:      "",
			expectedDeposits: []sdk.Coins{ukava(1e8).Add(usdx(1e8)...)},
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
