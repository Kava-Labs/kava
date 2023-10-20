package testutil

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/kava-labs/kava/app"
	types "github.com/kava-labs/kava/x/community/types"
	kavadisttypes "github.com/kava-labs/kava/x/kavadist/types"
)

func (suite *disableInflationTestSuite) TestStartCommunityFundConsolidation() {
	tests := []struct {
		name                   string
		initialFeePoolCoins    sdk.DecCoins
		initialKavadistBalance sdk.Coins
	}{
		{
			"basic test with both balances and dust",
			sdk.NewDecCoins(
				sdk.NewDecCoinFromDec("ukava", sdk.NewDecWithPrec(123456, 2)),
				sdk.NewDecCoinFromDec("usdx", sdk.NewDecWithPrec(654321, 3)),
			),
			sdk.NewCoins(
				sdk.NewInt64Coin("ukava", 10_000),
				sdk.NewInt64Coin("usdx", 10_000),
			),
		},
		{
			"empty x/distribution feepool",
			sdk.DecCoins(nil),
			sdk.NewCoins(
				sdk.NewInt64Coin("ukava", 10_000),
				sdk.NewInt64Coin("usdx", 10_000),
			),
		},
		{
			"empty x/kavadist balance",
			sdk.NewDecCoins(
				sdk.NewDecCoinFromDec("ukava", sdk.NewDecWithPrec(123456, 2)),
				sdk.NewDecCoinFromDec("usdx", sdk.NewDecWithPrec(654321, 3)),
			),
			sdk.Coins{},
		},
		{
			"both x/distribution feepool and x/kavadist balance empty",
			sdk.DecCoins(nil),
			sdk.Coins{},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			ak := suite.App.GetAccountKeeper()

			initialFeePool := distrtypes.FeePool{
				CommunityPool: tc.initialFeePoolCoins,
			}

			initialFeePoolCoins, initialFeePoolDust := initialFeePool.CommunityPool.TruncateDecimal()

			// More coins than initial feepool/communitypool
			fundCoins := sdk.NewCoins(
				sdk.NewInt64Coin("ukava", 10_000),
				sdk.NewInt64Coin("usdx", 10_000),
			)

			// Always fund x/distribution with enough coins to cover feepool
			err := suite.App.FundModuleAccount(
				suite.Ctx,
				distrtypes.ModuleName,
				fundCoins,
			)
			suite.NoError(err, "x/distribution account should be funded without error")

			err = suite.App.FundModuleAccount(
				suite.Ctx,
				kavadisttypes.ModuleName,
				tc.initialKavadistBalance,
			)
			suite.NoError(err, "x/kavadist account should be funded without error")

			suite.App.GetDistrKeeper().SetFeePool(suite.Ctx, initialFeePool)

			// Ensure the feepool was set before migration
			feePoolBefore := suite.App.GetDistrKeeper().GetFeePool(suite.Ctx)
			suite.Equal(initialFeePool, feePoolBefore, "initial feepool should be set")
			communityBalanceBefore := suite.App.GetCommunityKeeper().GetModuleAccountBalance(suite.Ctx)

			kavadistAcc := ak.GetModuleAccount(suite.Ctx, kavadisttypes.KavaDistMacc)
			kavaDistCoinsBefore := suite.App.GetBankKeeper().GetAllBalances(suite.Ctx, kavadistAcc.GetAddress())
			suite.Equal(
				tc.initialKavadistBalance,
				kavaDistCoinsBefore,
				"x/kavadist balance should be funded",
			)

			expectedKavaDistCoins := sdk.NewCoins(sdk.NewCoin("ukava", kavaDistCoinsBefore.AmountOf("ukava")))

			// -------------
			// Run upgrade

			params, found := suite.Keeper.GetParams(suite.Ctx)
			suite.Require().True(found)
			params.UpgradeTimeDisableInflation = suite.Ctx.BlockTime().Add(-time.Minute)
			suite.Keeper.SetParams(suite.Ctx, params)

			err = suite.Keeper.StartCommunityFundConsolidation(suite.Ctx)
			suite.NoError(err, "consolidation should not error")

			// -------------
			// Check results
			suite.Run("module balances after consolidation should moved", func() {
				feePoolAfter := suite.App.GetDistrKeeper().GetFeePool(suite.Ctx)
				suite.Equal(
					initialFeePoolDust,
					feePoolAfter.CommunityPool,
					"x/distribution community pool should be sent to x/community",
				)

				kavaDistCoinsAfter := suite.App.GetBankKeeper().GetAllBalances(suite.Ctx, kavadistAcc.GetAddress())
				suite.Equal(
					expectedKavaDistCoins,
					kavaDistCoinsAfter,
					"x/kavadist balance should ony contain ukava",
				)

				totalExpectedCommunityPoolCoins := communityBalanceBefore.
					Add(initialFeePoolCoins...).      // x/distribution fee pool
					Add(tc.initialKavadistBalance...) // x/kavadist module balance

				communityBalanceAfter := suite.App.GetCommunityKeeper().GetModuleAccountBalance(suite.Ctx)

				// Use .IsAllGTE to avoid types.Coins(nil) vs types.Coins{} mismatch
				suite.Truef(
					totalExpectedCommunityPoolCoins.IsAllGTE(communityBalanceAfter),
					"x/community balance should be increased by the truncated x/distribution community pool, got %s, expected %s",
					communityBalanceAfter,
					totalExpectedCommunityPoolCoins,
				)
			})

			suite.Run("bank transfer events should be emitted", func() {
				communityAcc := ak.GetModuleAccount(suite.Ctx, types.ModuleAccountName)
				distributionAcc := ak.GetModuleAccount(suite.Ctx, distrtypes.ModuleName)
				kavadistAcc := ak.GetModuleAccount(suite.Ctx, kavadisttypes.KavaDistMacc)

				events := suite.Ctx.EventManager().Events()

				suite.NoError(
					app.EventsContains(
						events,
						sdk.NewEvent(
							banktypes.EventTypeTransfer,
							sdk.NewAttribute(banktypes.AttributeKeyRecipient, communityAcc.GetAddress().String()),
							sdk.NewAttribute(banktypes.AttributeKeySender, distributionAcc.GetAddress().String()),
							sdk.NewAttribute(sdk.AttributeKeyAmount, initialFeePoolCoins.String()),
						),
					),
				)

				suite.NoError(
					app.EventsContains(
						events,
						sdk.NewEvent(
							banktypes.EventTypeTransfer,
							sdk.NewAttribute(banktypes.AttributeKeyRecipient, communityAcc.GetAddress().String()),
							sdk.NewAttribute(banktypes.AttributeKeySender, kavadistAcc.GetAddress().String()),
							sdk.NewAttribute(sdk.AttributeKeyAmount, kavaDistCoinsBefore.Sub(expectedKavaDistCoins...).String()),
						),
					),
				)
			})
		})
	}
}
