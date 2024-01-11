package keeper

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	earntypes "github.com/kava-labs/kava/x/earn/types"
	"github.com/kava-labs/kava/x/incentive/types"
	liquidtypes "github.com/kava-labs/kava/x/liquid/types"
)

const (
	SecondsPerYear = 31536000
)

// GetStakingAPR returns the total APR for staking and incentive rewards
func GetStakingAPR(ctx sdk.Context, k Keeper, params types.Params) (sdk.Dec, error) {
	// Get staking APR + incentive APR
	inflationRate := k.mintKeeper.GetMinter(ctx).Inflation
	communityTax := k.distrKeeper.GetCommunityTax(ctx)

	bondedTokens := k.stakingKeeper.TotalBondedTokens(ctx)
	circulatingSupply := k.bankKeeper.GetSupply(ctx, types.BondDenom)

	// Staking APR = (Inflation Rate * (1 - Community Tax)) / (Bonded Tokens / Circulating Supply)
	stakingAPR := inflationRate.
		Mul(sdk.OneDec().Sub(communityTax)).
		Quo(sdk.NewDecFromInt(bondedTokens).
			Quo(sdk.NewDecFromInt(circulatingSupply.Amount)))

	// Get incentive APR
	bkavaRewardPeriod, found := params.EarnRewardPeriods.GetMultiRewardPeriod(liquidtypes.DefaultDerivativeDenom)
	if !found {
		// No incentive rewards for bkava, only staking rewards
		return stakingAPR, nil
	}

	// Total amount of bkava in earn vaults, this may be lower than total bank
	// supply of bkava as some bkava may not be deposited in earn vaults
	totalEarnBkavaDeposited := sdk.ZeroInt()

	var iterErr error
	k.earnKeeper.IterateVaultRecords(ctx, func(record earntypes.VaultRecord) (stop bool) {
		if !k.liquidKeeper.IsDerivativeDenom(ctx, record.TotalShares.Denom) {
			return false
		}

		vaultValue, err := k.earnKeeper.GetVaultTotalValue(ctx, record.TotalShares.Denom)
		if err != nil {
			iterErr = err
			return false
		}

		totalEarnBkavaDeposited = totalEarnBkavaDeposited.Add(vaultValue.Amount)

		return false
	})

	if iterErr != nil {
		return sdk.ZeroDec(), iterErr
	}

	// Incentive APR = rewards per second * seconds per year / total supplied to earn vaults
	// Override collateral type to use "kava" instead of "bkava" when fetching
	incentiveAPY, err := GetAPYFromMultiRewardPeriod(ctx, k, types.BondDenom, bkavaRewardPeriod, totalEarnBkavaDeposited)
	if err != nil {
		return sdk.ZeroDec(), err
	}

	totalAPY := stakingAPR.Add(incentiveAPY)
	return totalAPY, nil
}

// GetAPYFromMultiRewardPeriod calculates the APY for a given MultiRewardPeriod
func GetAPYFromMultiRewardPeriod(
	ctx sdk.Context,
	k Keeper,
	collateralType string,
	rewardPeriod types.MultiRewardPeriod,
	totalSupply sdkmath.Int,
) (sdk.Dec, error) {
	if totalSupply.IsZero() {
		return sdk.ZeroDec(), nil
	}

	// Get USD value of collateral type
	collateralUSDValue, err := k.pricefeedKeeper.GetCurrentPrice(ctx, getMarketID(collateralType))
	if err != nil {
		return sdk.ZeroDec(), fmt.Errorf(
			"failed to get price for incentive collateralType %s with market ID %s: %w",
			collateralType, getMarketID(collateralType), err,
		)
	}

	// Total USD value of the collateral type total supply
	totalSupplyUSDValue := sdk.NewDecFromInt(totalSupply).Mul(collateralUSDValue.Price)

	totalUSDRewardsPerSecond := sdk.ZeroDec()

	// In many cases, RewardsPerSecond are assets that are different from the
	// CollateralType, so we need to use the USD value of CollateralType and
	// RewardsPerSecond to determine the APY.
	for _, reward := range rewardPeriod.RewardsPerSecond {
		// Get USD value of 1 unit of reward asset type, using TWAP
		rewardDenomUSDValue, err := k.pricefeedKeeper.GetCurrentPrice(ctx, getMarketID(reward.Denom))
		if err != nil {
			return sdk.ZeroDec(), fmt.Errorf("failed to get price for RewardsPerSecond asset %s: %w", reward.Denom, err)
		}

		rewardPerSecond := sdk.NewDecFromInt(reward.Amount).Mul(rewardDenomUSDValue.Price)
		totalUSDRewardsPerSecond = totalUSDRewardsPerSecond.Add(rewardPerSecond)
	}

	// APY = USD rewards per second * seconds per year / USD total supplied
	apy := totalUSDRewardsPerSecond.
		MulInt64(SecondsPerYear).
		Quo(totalSupplyUSDValue)

	return apy, nil
}

func getMarketID(denom string) string {
	// Rewrite denoms as pricefeed has different names for some assets,
	// e.g. "ukava" -> "kava", "erc20/multichain/usdc" -> "usdc"
	// bkava is not included as it is handled separately

	// TODO: Replace hardcoded conversion with possible params set somewhere
	// to be more flexible. E.g. a map of denoms to pricefeed market denoms in
	// pricefeed params.
	switch denom {
	case types.BondDenom:
		denom = "kava"
	case "erc20/multichain/usdc":
		denom = "usdc"
	case "erc20/multichain/usdt":
		denom = "usdt"
	case "erc20/multichain/dai":
		denom = "dai"
	}

	return fmt.Sprintf("%s:usd:30", denom)
}
