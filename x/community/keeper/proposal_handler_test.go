package keeper_test

import (
	"fmt"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/kava-labs/kava/app"
	cdpkeeper "github.com/kava-labs/kava/x/cdp/keeper"
	"github.com/kava-labs/kava/x/community/keeper"
	"github.com/kava-labs/kava/x/community/testutil"
	"github.com/kava-labs/kava/x/community/types"
	hardkeeper "github.com/kava-labs/kava/x/hard/keeper"
	hardtypes "github.com/kava-labs/kava/x/hard/types"
	pricefeedtypes "github.com/kava-labs/kava/x/pricefeed/types"
)

const chainID = "kavatest_2221-1"

func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }
func ukava(amt int64) sdk.Coins {
	return sdk.NewCoins(c("ukava", amt))
}

func usdx(amt int64) sdk.Coins {
	return sdk.NewCoins(c("usdx", amt))
}

func otherdenom(amt int64) sdk.Coins {
	return sdk.NewCoins(c("other-denom", amt))
}

type proposalTestSuite struct {
	suite.Suite

	App         app.TestApp
	Ctx         sdk.Context
	Keeper      keeper.Keeper
	MaccAddress sdk.AccAddress

	cdpKeeper  cdpkeeper.Keeper
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

	// Set UpgradeTimeDisableInflation to far future to not influence module
	// account balances
	params := types.Params{
		UpgradeTimeDisableInflation: time.Now().Add(100000 * time.Hour),
		StakingRewardsPerSecond:     sdkmath.LegacyNewDec(0),
	}
	communityGs := types.NewGenesisState(params, types.DefaultStakingRewardsState())

	tApp.InitializeFromGenesisStatesWithTimeAndChainID(
		genTime, chainID,
		app.GenesisState{hardtypes.ModuleName: tApp.AppCodec().MustMarshalJSON(&hardGS)},
		app.GenesisState{pricefeedtypes.ModuleName: tApp.AppCodec().MustMarshalJSON(&pricefeedGS)},
		app.GenesisState{types.ModuleName: tApp.AppCodec().MustMarshalJSON(&communityGs)},
		testutil.NewCDPGenState(tApp.AppCodec(), "ukava", "kava", sdk.NewDec(2)),
	)

	suite.App = tApp
	suite.Ctx = ctx
	suite.Keeper = tApp.GetCommunityKeeper()
	suite.MaccAddress = tApp.GetAccountKeeper().GetModuleAddress(types.ModuleAccountName)
	suite.cdpKeeper = suite.App.GetCDPKeeper()
	suite.hardKeeper = suite.App.GetHardKeeper()

	// give the community pool some funds
	// ukava
	suite.FundCommunityPool(ukava(2e10))
	// usdx
	suite.FundCommunityPool(usdx(2e10))
	// other-denom
	suite.FundCommunityPool(otherdenom(1e10))
}

func (suite *proposalTestSuite) NextBlock() {
	newTime := suite.Ctx.BlockTime().Add(6 * time.Second)
	newHeight := suite.Ctx.BlockHeight() + 1

	suite.App.EndBlocker(suite.Ctx, abcitypes.RequestEndBlock{})
	suite.Ctx = suite.Ctx.WithBlockTime(newTime).WithBlockHeight(newHeight).WithChainID(chainID)
	suite.App.BeginBlocker(suite.Ctx, abcitypes.RequestBeginBlock{})
}

func (suite *proposalTestSuite) FundCommunityPool(coins sdk.Coins) {
	// mint to ephemeral account
	ephemeralAcc := app.RandomAddress()
	suite.NoError(suite.App.FundAccount(suite.Ctx, ephemeralAcc, coins))
	// fund community pool with newly minted coins
	suite.NoError(suite.App.GetDistrKeeper().FundCommunityPool(suite.Ctx, coins, ephemeralAcc))
}

func (suite *proposalTestSuite) GetCommunityPoolBalance() sdk.Coins {
	ak := suite.App.GetAccountKeeper()
	bk := suite.App.GetBankKeeper()

	addr := ak.GetModuleAddress(types.ModuleAccountName)

	// Return x/community module account balance, no longer using x/distribution community pool
	return bk.GetAllBalances(suite.Ctx, addr)
}

func (suite *proposalTestSuite) CheckCommunityPoolBalance(expected sdk.Coins) {
	actual := suite.GetCommunityPoolBalance()
	// check that balance is expected
	suite.True(expected.IsEqual(actual), fmt.Sprintf("unexpected balance in community pool\nexpected: %s\nactual: %s", expected, actual))
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
			expectedErr:      "community pool does not have sufficient coins to distribute",
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

func (suite *proposalTestSuite) TestCommunityLendWithdrawProposal() {
	testCases := []struct {
		name               string
		initialDeposit     sdk.Coins
		proposals          []*types.CommunityPoolLendWithdrawProposal
		expectedErr        string
		expectedWithdrawal sdk.Coins
	}{
		{
			// in the week it would take a proposal to pass, the position would have grown
			// to withdraw the entire position, it'd be safest to set a very high withdraw
			name:           "valid - requesting withdrawal of more than total will withdraw all",
			initialDeposit: ukava(1e9),
			proposals: []*types.CommunityPoolLendWithdrawProposal{
				{Amount: ukava(1e12)},
			},
			expectedErr:        "",
			expectedWithdrawal: ukava(1e9),
		},
		{
			name:           "valid - single proposal, single denom, full withdrawal",
			initialDeposit: ukava(1e9),
			proposals: []*types.CommunityPoolLendWithdrawProposal{
				{Amount: ukava(1e9)},
			},
			expectedErr:        "",
			expectedWithdrawal: ukava(1e9),
		},
		{
			name:           "valid - single proposal, multiple denoms, full withdrawal",
			initialDeposit: ukava(1e9).Add(usdx(1e9)...),
			proposals: []*types.CommunityPoolLendWithdrawProposal{
				{Amount: ukava(1e9).Add(usdx(1e9)...)},
			},
			expectedErr:        "",
			expectedWithdrawal: ukava(1e9).Add(usdx(1e9)...),
		},
		{
			name:           "valid - single proposal, partial withdrawal",
			initialDeposit: ukava(1e9).Add(usdx(1e9)...),
			proposals: []*types.CommunityPoolLendWithdrawProposal{
				{Amount: ukava(1e8).Add(usdx(1e9)...)},
			},
			expectedErr:        "",
			expectedWithdrawal: ukava(1e8).Add(usdx(1e9)...),
		},
		{
			name:           "valid - multiple proposals, full withdrawal",
			initialDeposit: ukava(1e9).Add(usdx(1e9)...),
			proposals: []*types.CommunityPoolLendWithdrawProposal{
				{Amount: ukava(1e9)},
				{Amount: usdx(1e9)},
			},
			expectedErr:        "",
			expectedWithdrawal: ukava(1e9).Add(usdx(1e9)...),
		},
		{
			name:           "valid - multiple proposals, partial withdrawal",
			initialDeposit: ukava(1e9).Add(usdx(1e9)...),
			proposals: []*types.CommunityPoolLendWithdrawProposal{
				{Amount: ukava(1e8)},
				{Amount: usdx(1e8)},
			},
			expectedErr:        "",
			expectedWithdrawal: ukava(1e8).Add(usdx(1e8)...),
		},
		{
			name:           "invalid - nonexistent position, has no deposits",
			initialDeposit: sdk.NewCoins(),
			proposals: []*types.CommunityPoolLendWithdrawProposal{
				{Amount: ukava(1e8)},
			},
			expectedErr:        "deposit not found",
			expectedWithdrawal: sdk.NewCoins(),
		},
		{
			name:           "invalid - nonexistent position, has deposits of different denom",
			initialDeposit: ukava(1e8),
			proposals: []*types.CommunityPoolLendWithdrawProposal{
				{Amount: usdx(1e8)},
			},
			expectedErr:        "no coins of this type deposited",
			expectedWithdrawal: sdk.NewCoins(),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			// Disable minting, so that the community pool balance doesn't change
			// during the test - this is because staking denom is "ukava" and no
			// longer "stake" which has an initial and changing balance instead
			// of just 0
			suite.App.SetInflation(suite.Ctx, sdk.ZeroDec())

			// setup initial deposit
			if !tc.initialDeposit.IsZero() {
				deposit := types.NewCommunityPoolLendDepositProposal("initial deposit", "has coins", tc.initialDeposit)
				err := keeper.HandleCommunityPoolLendDepositProposal(suite.Ctx, suite.Keeper, deposit)
				suite.NoError(err, "unexpected error while seeding lend deposit")
			}

			beforeBalance := suite.GetCommunityPoolBalance()

			// run the proposals
			for i, proposal := range tc.proposals {
				fmt.Println("submitting proposal ", i, " ", suite.Ctx.ChainID())
				err := keeper.HandleCommunityPoolLendWithdrawProposal(suite.Ctx, suite.Keeper, proposal)
				if tc.expectedErr == "" {
					suite.NoError(err)
				} else {
					suite.ErrorContains(err, tc.expectedErr)
				}

				suite.NextBlock()
			}

			// expect funds to be removed from hard deposit
			expectedRemaining := tc.initialDeposit.Sub(tc.expectedWithdrawal...)
			deposits := suite.hardKeeper.GetDepositsByUser(suite.Ctx, suite.MaccAddress)
			if expectedRemaining.IsZero() {
				suite.Len(deposits, 0, "expected all deposits to be withdrawn")
			} else {
				suite.Len(deposits, 1, "expected user to have remaining deposit")
				suite.Equal(expectedRemaining, deposits[0].Amount)
			}

			// expect funds to be distributed back to community pool
			suite.CheckCommunityPoolBalance(beforeBalance.Add(tc.expectedWithdrawal...))
		})
	}
}

// expectation: funds in the community module will be used to repay cdps.
// if collateral is returned, it stays in the community module.
func (suite *proposalTestSuite) TestCommunityCDPRepayDebtProposal() {
	initialModuleFunds := ukava(2e10).Add(otherdenom(1e9)...)
	collateralType := "kava-a"
	type debt struct {
		collateral sdk.Coin
		principal  sdk.Coin
	}
	testcases := []struct {
		name           string
		initialDebt    *debt
		proposal       *types.CommunityCDPRepayDebtProposal
		expectedErr    string
		expectedRepaid sdk.Coin
	}{
		{
			name:        "valid - paid in full",
			initialDebt: &debt{c("ukava", 1e10), c("usdx", 1e9)},
			proposal: types.NewCommunityCDPRepayDebtProposal(
				"repaying my debts in full",
				"title says it all",
				collateralType,
				c("usdx", 1e9),
			),
			expectedErr:    "",
			expectedRepaid: c("usdx", 1e9),
		},
		{
			name:        "valid - partial payment",
			initialDebt: &debt{c("ukava", 1e10), c("usdx", 1e9)},
			proposal: types.NewCommunityCDPRepayDebtProposal(
				"title goes here",
				"description goes here",
				collateralType,
				c("usdx", 1e8),
			),
			expectedErr:    "",
			expectedRepaid: c("usdx", 1e8),
		},
		{
			name:        "invalid - insufficient funds",
			initialDebt: &debt{c("ukava", 1e10), c("usdx", 1e9)},
			proposal: types.NewCommunityCDPRepayDebtProposal(
				"title goes here",
				"description goes here",
				collateralType,
				c("usdx", 1e10), // <-- more usdx than we have
			),
			expectedErr:    "insufficient balance",
			expectedRepaid: c("usdx", 0),
		},
	}

	for _, tc := range testcases {
		suite.Run(tc.name, func() {
			var err error
			suite.SetupTest()

			// setup the community module with some initial funds
			err = suite.App.FundModuleAccount(suite.Ctx, types.ModuleAccountName, initialModuleFunds)
			suite.NoError(err, "failed to initially fund module account for cdp creation")

			// setup initial debt position
			err = suite.cdpKeeper.AddCdp(suite.Ctx, suite.MaccAddress, tc.initialDebt.collateral, tc.initialDebt.principal, collateralType)
			suite.NoError(err, "unexpected error while creating initial cdp")

			balanceBefore := suite.Keeper.GetModuleAccountBalance(suite.Ctx)

			// submit proposal
			err = keeper.HandleCommunityCDPRepayDebtProposal(suite.Ctx, suite.Keeper, tc.proposal)
			if tc.expectedErr == "" {
				suite.NoError(err)
			} else {
				suite.ErrorContains(err, tc.expectedErr)
			}
			suite.NextBlock()

			cdps := suite.cdpKeeper.GetAllCdpsByCollateralType(suite.Ctx, collateralType)
			expectedRemainingPrincipal := tc.initialDebt.principal.Sub(tc.expectedRepaid)
			fullyRepaid := expectedRemainingPrincipal.IsZero()

			// expect repayment funds to be deducted from community module account
			expectedModuleBalance := balanceBefore.Sub(tc.expectedRepaid)
			// when fully repaid, the position is closed and collateral is returned.
			if fullyRepaid {
				suite.Len(cdps, 0, "expected position to have been closed on payment")
				// expect balance to include recouped collateral
				expectedModuleBalance = expectedModuleBalance.Add(tc.initialDebt.collateral)
			} else {
				suite.Len(cdps, 1, "expected debt position to remain open")
				suite.Equal(suite.MaccAddress, cdps[0].Owner, "sanity check: unexpected owner")
				// check the remaining principle on the cdp
				suite.Equal(expectedRemainingPrincipal, cdps[0].Principal)
			}

			// verify the balance changed as expected
			moduleBalanceAfter := suite.Keeper.GetModuleAccountBalance(suite.Ctx)
			suite.True(expectedModuleBalance.IsEqual(moduleBalanceAfter), "module balance changed unexpectedly")
		})
	}
}

// expectation: funds in the community module used as cdp collateral will be
// withdrawn and stays in the community module.
func (suite *proposalTestSuite) TestCommunityCDPWithdrawCollateralProposal() {
	initialModuleFunds := ukava(2e10).Add(otherdenom(1e9)...)
	collateralType := "kava-a"
	type debt struct {
		collateral sdk.Coin
		principal  sdk.Coin
	}
	testcases := []struct {
		name              string
		initialDebt       *debt
		proposal          *types.CommunityCDPWithdrawCollateralProposal
		expectedErr       string
		expectedWithdrawn sdk.Coin
	}{
		{
			name: "valid - withdrawing max collateral",
			initialDebt: &debt{
				c("ukava", 1e10),
				c("usdx", 1e9),
			},
			proposal: types.NewCommunityCDPWithdrawCollateralProposal(
				"withdrawing max collateral",
				"i might get liquidated",
				collateralType,
				c("ukava", 8e9-1), // Withdraw all collateral except 2*principal-1 amount
			),
			expectedErr:       "",
			expectedWithdrawn: c("ukava", 8e9-1),
		},
		{
			name: "valid - withdrawing partial collateral",
			initialDebt: &debt{
				c("ukava", 1e10),
				c("usdx", 1e9),
			},
			proposal: types.NewCommunityCDPWithdrawCollateralProposal(
				"title goes here",
				"description goes here",
				collateralType,
				c("ukava", 1e9),
			),
			expectedErr:       "",
			expectedWithdrawn: c("ukava", 1e9),
		},
		{
			name: "invalid - withdrawing too much collateral",
			initialDebt: &debt{
				c("ukava", 1e10),
				c("usdx", 1e9),
			},
			proposal: types.NewCommunityCDPWithdrawCollateralProposal(
				"title goes here",
				"description goes here",
				collateralType,
				c("ukava", 9e9), // <-- would be under collateralized
			),
			expectedErr:       "proposed collateral ratio is below liquidation ratio",
			expectedWithdrawn: c("ukava", 0),
		},
	}

	for _, tc := range testcases {
		suite.Run(tc.name, func() {
			var err error
			suite.SetupTest()

			// setup the community module with some initial funds
			err = suite.App.FundModuleAccount(suite.Ctx, types.ModuleAccountName, initialModuleFunds)
			suite.NoError(err, "failed to initially fund module account for cdp creation")

			// setup initial debt position
			err = suite.cdpKeeper.AddCdp(suite.Ctx, suite.MaccAddress, tc.initialDebt.collateral, tc.initialDebt.principal, collateralType)
			suite.NoError(err, "unexpected error while creating initial cdp")

			balanceBefore := suite.Keeper.GetModuleAccountBalance(suite.Ctx)

			// submit proposal
			err = keeper.HandleCommunityCDPWithdrawCollateralProposal(suite.Ctx, suite.Keeper, tc.proposal)
			if tc.expectedErr == "" {
				suite.NoError(err)
			} else {
				suite.Require().ErrorContains(err, tc.expectedErr)
			}
			suite.NextBlock()

			cdps := suite.cdpKeeper.GetAllCdpsByCollateralType(suite.Ctx, collateralType)
			expectedRemainingCollateral := tc.initialDebt.collateral.Sub(tc.expectedWithdrawn)

			// expect withdrawn funds to add to community module account
			expectedModuleBalance := balanceBefore.Add(tc.expectedWithdrawn)

			suite.Len(cdps, 1, "expected debt position to remain open")
			suite.Equal(suite.MaccAddress, cdps[0].Owner, "sanity check: unexpected owner")
			// check the remaining principle on the cdp
			suite.Equal(expectedRemainingCollateral, cdps[0].Collateral)

			// verify the balance changed as expected
			moduleBalanceAfter := suite.Keeper.GetModuleAccountBalance(suite.Ctx)
			suite.True(expectedModuleBalance.IsEqual(moduleBalanceAfter), "module balance changed unexpectedly")
		})
	}
}
