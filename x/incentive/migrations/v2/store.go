package v2

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/incentive/types"
)

// MigrateParams performs in-place migrations from kava 10 to kava 11.
// The migration includes:
//
// 1. Add rewards for bkava liquid staking = 10% APR
// 2. Add rewards for USDC = 10% APR
// 3. Change APR for vanilla kava staking from 23% --> 20%
// 4. Remove lockup periods from claims (going forward)
// 5. Remove HARD and SWP delegation rewards for vanilla staking
func MigrateParams(ctx sdk.Context, paramstore types.ParamSubspace) error {
	if !paramstore.HasKeyTable() {
		paramstore = paramstore.WithKeyTable(types.ParamKeyTable())
	}

	var oldParams v1Params
	paramstore.GetParamSet(ctx, &oldParams)

	params := types.Params{
		USDXMintingRewardPeriods: oldParams.USDXMintingRewardPeriods,
		HardSupplyRewardPeriods:  oldParams.HardSupplyRewardPeriods,
		HardBorrowRewardPeriods:  oldParams.HardBorrowRewardPeriods,
		DelegatorRewardPeriods:   oldParams.DelegatorRewardPeriods,
		SwapRewardPeriods:        oldParams.SwapRewardPeriods,
		ClaimMultipliers:         oldParams.ClaimMultipliers,
		ClaimEnd:                 oldParams.ClaimEnd,
		SavingsRewardPeriods:     oldParams.SavingsRewardPeriods,
		EarnRewardPeriods:        nil,
	}

	periodStart := time.Date(2022, 10, 26, 15, 0, 0, 0, time.UTC)
	periodEnd := time.Date(2024, 10, 26, 15, 0, 0, 0, time.UTC)

	params.EarnRewardPeriods = types.MultiRewardPeriods{
		// 1. Add rewards for bkava liquid staking = 10% APR
		types.NewMultiRewardPeriod(
			true,
			"bkava",
			periodStart,
			periodEnd,
			sdk.NewCoins(
				// 0.1902587519 KAVA == 190258.752ukava
				sdk.NewCoin("ukava", sdk.NewInt(190258)),
			),
		),
		// 2. Add rewards for USDC = 10% APR
		types.NewMultiRewardPeriod(
			true,
			"erc20/multichain/usdc",
			periodStart,
			periodEnd,
			sdk.NewCoins(
				// 0.005284965331 USDC == 5284.965331 (6 decimals!)
				sdk.NewCoin("ukava", sdk.NewInt(5284)),
			),
		),
	}

	// Update hard supply rewards
	for i, period := range params.HardSupplyRewardPeriods {
		if period.CollateralType == "usdx" {
			params.HardSupplyRewardPeriods[i] = types.NewMultiRewardPeriod(
				true,
				period.CollateralType,
				period.Start,
				period.End,
				sdk.NewCoins(
					// Same hard rewards as before
					sdk.NewCoin("hard", period.RewardsPerSecond.AmountOf("hard")),
					// 0.4735328936ukava == 473532.8936ukava
					sdk.NewCoin("ukava", sdk.NewInt(473532)),
				),
			)
		}

		// ATOM
		// https://api.data.kava.io/ibc/apps/transfer/v1/denom_traces/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2
		if period.CollateralType == "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2" {
			params.HardSupplyRewardPeriods[i] = types.NewMultiRewardPeriod(
				true,
				period.CollateralType,
				period.Start,
				period.End,
				sdk.NewCoins(
					// Same hard rewards as before
					sdk.NewCoin("hard", period.RewardsPerSecond.AmountOf("hard")),
					// 0.01109842719ukava == 11098.42719ukava
					sdk.NewCoin("ukava", sdk.NewInt(11098)),
				),
			)
		}
	}

	// New multiplier with 0 lockup and full rewards factor
	newMultiplier := types.NewMultiplier("large", 0, sdk.OneDec())

	// 4. Remove lockup periods from claims (going forward)
	// All claim multipliers have 0 lockup period
	params.ClaimMultipliers = types.MultipliersPerDenoms{
		{
			Denom: "hard",
			Multipliers: types.Multipliers{
				newMultiplier,
			},
		},
		{
			Denom: "ukava",
			Multipliers: types.Multipliers{
				newMultiplier,
			},
		},
		{
			Denom: "swp",
			Multipliers: types.Multipliers{
				newMultiplier,
			},
		},
		{
			Denom: "ibc/799FDD409719A1122586A629AE8FCA17380351A51C1F47A80A1B8E7F2A491098",
			Multipliers: types.Multipliers{
				newMultiplier,
			},
		},
	}

	// 5. Remove HARD and SWP rewards for vanilla staking
	params.DelegatorRewardPeriods = types.MultiRewardPeriods{}

	if err := params.Validate(); err != nil {
		return err
	}

	paramstore.SetParamSet(ctx, &params)
	return nil
}
