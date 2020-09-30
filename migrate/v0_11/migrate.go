package v0_11

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	v0_11bep3 "github.com/kava-labs/kava/x/bep3/legacy/v0_11"
	v0_9bep3 "github.com/kava-labs/kava/x/bep3/legacy/v0_9"
	v0_11harvest "github.com/kava-labs/kava/x/harvest"
	v0_11incentive "github.com/kava-labs/kava/x/incentive"
	v0_9incentive "github.com/kava-labs/kava/x/incentive/legacy/v0_9"
	v0_11pricefeed "github.com/kava-labs/kava/x/pricefeed"
	v0_9pricefeed "github.com/kava-labs/kava/x/pricefeed/legacy/v0_9"
)

// MigrateBep3 migrates from a v0.9 (or v0.10) bep3 genesis state to a v0.11 bep3 genesis state
func MigrateBep3(oldGenState v0_9bep3.GenesisState) v0_11bep3.GenesisState {
	var assetParams v0_11bep3.AssetParams
	var assetSupplies v0_11bep3.AssetSupplies
	v0_9Params := oldGenState.Params

	for _, asset := range v0_9Params.SupportedAssets {
		v10AssetParam := v0_11bep3.AssetParam{
			Active:        asset.Active,
			Denom:         asset.Denom,
			CoinID:        asset.CoinID,
			DeputyAddress: v0_9Params.BnbDeputyAddress,
			FixedFee:      v0_9Params.BnbDeputyFixedFee,
			MinSwapAmount: sdk.OneInt(), // set min swap to one - prevents accounts that hold zero bnb from creating spam txs
			MaxSwapAmount: v0_9Params.MaxAmount,
			MinBlockLock:  v0_9Params.MinBlockLock,
			MaxBlockLock:  v0_9Params.MaxBlockLock,
			SupplyLimit: v0_11bep3.SupplyLimit{
				Limit:          asset.Limit,
				TimeLimited:    false,
				TimePeriod:     time.Duration(0),
				TimeBasedLimit: sdk.ZeroInt(),
			},
		}
		assetParams = append(assetParams, v10AssetParam)
	}
	for _, supply := range oldGenState.AssetSupplies {
		newSupply := v0_11bep3.NewAssetSupply(supply.IncomingSupply, supply.OutgoingSupply, supply.CurrentSupply, sdk.NewCoin(supply.CurrentSupply.Denom, sdk.ZeroInt()), time.Duration(0))
		assetSupplies = append(assetSupplies, newSupply)
	}
	var swaps v0_11bep3.AtomicSwaps
	for _, oldSwap := range oldGenState.AtomicSwaps {
		newSwap := v0_11bep3.AtomicSwap{
			Amount:              oldSwap.Amount,
			RandomNumberHash:    oldSwap.RandomNumberHash,
			ExpireHeight:        oldSwap.ExpireHeight,
			Timestamp:           oldSwap.Timestamp,
			Sender:              oldSwap.Sender,
			Recipient:           oldSwap.Recipient,
			SenderOtherChain:    oldSwap.SenderOtherChain,
			RecipientOtherChain: oldSwap.RecipientOtherChain,
			ClosedBlock:         oldSwap.ClosedBlock,
			Status:              v0_11bep3.SwapStatus(oldSwap.Status),
			CrossChain:          oldSwap.CrossChain,
			Direction:           v0_11bep3.SwapDirection(oldSwap.Direction),
		}
		swaps = append(swaps, newSwap)
	}
	return v0_11bep3.GenesisState{
		Params: v0_11bep3.Params{
			AssetParams: assetParams},
		AtomicSwaps:       swaps,
		Supplies:          assetSupplies,
		PreviousBlockTime: v0_11bep3.DefaultPreviousBlockTime,
	}
}

// MigrateHarvest initializes the harvest genesis state for kava-4
func MigrateHarvest() v0_11harvest.GenesisState {
	// total HARD per second for lps (week one): 633761
	// HARD per second for delegators (week one): 1267522
	incentiveGoLiveDate := time.Date(2020, 10, 16, 14, 0, 0, 0, time.UTC)
	incentiveEndDate := time.Date(2024, 10, 16, 14, 0, 0, 0, time.UTC)
	claimEndDate := time.Date(2025, 10, 16, 14, 0, 0, 0, time.UTC)
	harvestGS := v0_11harvest.NewGenesisState(v0_11harvest.NewParams(
		true,
		v0_11harvest.DistributionSchedules{
			v0_11harvest.NewDistributionSchedule(true, "usdx", incentiveGoLiveDate, incentiveEndDate, sdk.NewCoin("hard", sdk.NewInt(310543)), claimEndDate, v0_11harvest.Multipliers{v0_11harvest.NewMultiplier(v0_11harvest.Small, 1, sdk.MustNewDecFromStr("0.33")), v0_11harvest.NewMultiplier(v0_11harvest.Large, 12, sdk.OneDec())}),
			v0_11harvest.NewDistributionSchedule(true, "hard", incentiveGoLiveDate, incentiveEndDate, sdk.NewCoin("hard", sdk.NewInt(285193)), claimEndDate, v0_11harvest.Multipliers{v0_11harvest.NewMultiplier(v0_11harvest.Small, 1, sdk.MustNewDecFromStr("0.33")), v0_11harvest.NewMultiplier(v0_11harvest.Large, 12, sdk.OneDec())}),
			v0_11harvest.NewDistributionSchedule(true, "bnb", incentiveGoLiveDate, incentiveEndDate, sdk.NewCoin("hard", sdk.NewInt(12675)), claimEndDate, v0_11harvest.Multipliers{v0_11harvest.NewMultiplier(v0_11harvest.Small, 1, sdk.MustNewDecFromStr("0.33")), v0_11harvest.NewMultiplier(v0_11harvest.Large, 12, sdk.OneDec())}),
			v0_11harvest.NewDistributionSchedule(true, "ukava", incentiveGoLiveDate, incentiveEndDate, sdk.NewCoin("hard", sdk.NewInt(25350)), claimEndDate, v0_11harvest.Multipliers{v0_11harvest.NewMultiplier(v0_11harvest.Small, 1, sdk.MustNewDecFromStr("0.33")), v0_11harvest.NewMultiplier(v0_11harvest.Large, 12, sdk.OneDec())}),
		},
		v0_11harvest.DelegatorDistributionSchedules{v0_11harvest.NewDelegatorDistributionSchedule(
			v0_11harvest.NewDistributionSchedule(true, "ukava", incentiveGoLiveDate, incentiveEndDate, sdk.NewCoin("hard", sdk.NewInt(1267522)), claimEndDate, v0_11harvest.Multipliers{v0_11harvest.NewMultiplier(v0_11harvest.Small, 1, sdk.MustNewDecFromStr("0.33")), v0_11harvest.NewMultiplier(v0_11harvest.Large, 12, sdk.OneDec())}),
			time.Hour*24,
		),
		},
	), v0_11harvest.DefaultPreviousBlockTime, v0_11harvest.DefaultDistributionTimes)
	return harvestGS
}

// MigrateIncentive migrates from a v0.9 (or v0.10) incentive genesis state to a v0.11 incentive genesis state
func MigrateIncentive(oldGenState v0_9incentive.GenesisState) v0_11incentive.GenesisState {
	var newRewards v0_11incentive.Rewards
	var newRewardPeriods v0_11incentive.RewardPeriods
	var newClaimPeriods v0_11incentive.ClaimPeriods
	var newClaims v0_11incentive.Claims
	var newClaimPeriodIds v0_11incentive.GenesisClaimPeriodIDs

	newMultiplier := v0_11incentive.NewMultiplier(v0_11incentive.Large, 12, sdk.OneDec())
	smallMultiplier := v0_11incentive.NewMultiplier(v0_11incentive.Small, 1, sdk.MustNewDecFromStr("0.25"))

	for _, oldReward := range oldGenState.Params.Rewards {
		newReward := v0_11incentive.NewReward(oldReward.Active, oldReward.Denom+"-a", oldReward.AvailableRewards, oldReward.Duration, v0_11incentive.Multipliers{smallMultiplier, newMultiplier}, oldReward.ClaimDuration)
		newRewards = append(newRewards, newReward)
	}
	newParams := v0_11incentive.NewParams(true, newRewards)

	for _, oldRewardPeriod := range oldGenState.RewardPeriods {

		newRewardPeriod := v0_11incentive.NewRewardPeriod(oldRewardPeriod.Denom+"-a", oldRewardPeriod.Start, oldRewardPeriod.End, oldRewardPeriod.Reward, oldRewardPeriod.ClaimEnd, v0_11incentive.Multipliers{smallMultiplier, newMultiplier})
		newRewardPeriods = append(newRewardPeriods, newRewardPeriod)
	}

	for _, oldClaimPeriod := range oldGenState.ClaimPeriods {
		newClaimPeriod := v0_11incentive.NewClaimPeriod(oldClaimPeriod.Denom+"-a", oldClaimPeriod.ID, oldClaimPeriod.End, v0_11incentive.Multipliers{smallMultiplier, newMultiplier})
		newClaimPeriods = append(newClaimPeriods, newClaimPeriod)
	}

	for _, oldClaim := range oldGenState.Claims {
		newClaim := v0_11incentive.NewClaim(oldClaim.Owner, oldClaim.Reward, oldClaim.Denom+"-a", oldClaim.ClaimPeriodID)
		newClaims = append(newClaims, newClaim)
	}

	for _, oldClaimPeriodID := range oldGenState.NextClaimPeriodIDs {
		newClaimPeriodID := v0_11incentive.GenesisClaimPeriodID{
			CollateralType: oldClaimPeriodID.Denom + "-a",
			ID:             oldClaimPeriodID.ID,
		}
		newClaimPeriodIds = append(newClaimPeriodIds, newClaimPeriodID)
	}

	return v0_11incentive.NewGenesisState(newParams, oldGenState.PreviousBlockTime, newRewardPeriods, newClaimPeriods, newClaims, newClaimPeriodIds)
}

// MigratePricefeed migrates from a v0.9 (or v0.10) pricefeed genesis state to a v0.11 pricefeed genesis state
func MigratePricefeed(oldGenState v0_9pricefeed.GenesisState) v0_11pricefeed.GenesisState {
	var newMarkets v0_11pricefeed.Markets
	var newPostedPrices v0_11pricefeed.PostedPrices
	var oracles []sdk.AccAddress

	for _, market := range oldGenState.Params.Markets {
		newMarket := v0_11pricefeed.NewMarket(market.MarketID, market.BaseAsset, market.QuoteAsset, market.Oracles, market.Active)
		newMarkets = append(newMarkets, newMarket)
		oracles = market.Oracles
	}
	// ------- add btc, xrp, busd markets --------
	btcSpotMarket := v0_11pricefeed.NewMarket("btc:usd", "btc", "usd", oracles, true)
	btcLiquidationMarket := v0_11pricefeed.NewMarket("btc:usd:30", "btc", "usd", oracles, true)
	xrpSpotMarket := v0_11pricefeed.NewMarket("xrp:usd", "xrp", "usd", oracles, true)
	xrpLiquidationMarket := v0_11pricefeed.NewMarket("xrp:usd:30", "xrp", "usd", oracles, true)
	busdSpotMarket := v0_11pricefeed.NewMarket("busd:usd", "busd", "usd", oracles, true)
	busdLiquidationMarket := v0_11pricefeed.NewMarket("busd:usd:30", "busd", "usd", oracles, true)
	newMarkets = append(newMarkets, btcSpotMarket, btcLiquidationMarket, xrpSpotMarket, xrpLiquidationMarket, busdSpotMarket, busdLiquidationMarket)

	for _, price := range oldGenState.PostedPrices {
		newPrice := v0_11pricefeed.NewPostedPrice(price.MarketID, price.OracleAddress, price.Price, price.Expiry)
		newPostedPrices = append(newPostedPrices, newPrice)
	}
	newParams := v0_11pricefeed.NewParams(newMarkets)

	return v0_11pricefeed.NewGenesisState(newParams, newPostedPrices)
}
