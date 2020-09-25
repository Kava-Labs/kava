package v0_11

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	v0_11bep3 "github.com/kava-labs/kava/x/bep3/legacy/v0_11"
	v0_9bep3 "github.com/kava-labs/kava/x/bep3/legacy/v0_9"
	v0_11harvest "github.com/kava-labs/kava/x/hvt"
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
	// total HARD per second for lps: 634196
	// HARD per second for delegators: 1268392
	incentiveGoLiveDate := time.Date(2020, 10, 17, 14, 0, 0, 0, time.UTC)
	incentiveEndDate := time.Date(2024, 10, 17, 14, 0, 0, 0, time.UTC)
	claimEndDate := time.Date(2025, 10, 17, 14, 0, 0, 0, time.UTC)
	harvestGS := v0_11harvest.NewGenesisState(v0_11harvest.NewParams(
		true,
		v0_11harvest.DistributionSchedules{
			v0_11harvest.NewDistributionSchedule(true, "usdx", incentiveGoLiveDate, incentiveEndDate, sdk.NewCoin("hard", sdk.NewInt(253678)), claimEndDate, v0_11harvest.Multipliers{v0_11harvest.NewMultiplier(v0_11harvest.Small, 0, sdk.MustNewDecFromStr("0.33")), v0_11harvest.NewMultiplier(v0_11harvest.Medium, 6, sdk.MustNewDecFromStr("0.5")), v0_11harvest.NewMultiplier(v0_11harvest.Large, 24, sdk.OneDec())}),
			v0_11harvest.NewDistributionSchedule(true, "hard", incentiveGoLiveDate, incentiveEndDate, sdk.NewCoin("hard", sdk.NewInt(253678)), claimEndDate, v0_11harvest.Multipliers{v0_11harvest.NewMultiplier(v0_11harvest.Small, 0, sdk.MustNewDecFromStr("0.33")), v0_11harvest.NewMultiplier(v0_11harvest.Medium, 6, sdk.MustNewDecFromStr("0.5")), v0_11harvest.NewMultiplier(v0_11harvest.Large, 24, sdk.OneDec())}),
			v0_11harvest.NewDistributionSchedule(true, "btcb", incentiveGoLiveDate, incentiveEndDate, sdk.NewCoin("hard", sdk.NewInt(63420)), claimEndDate, v0_11harvest.Multipliers{v0_11harvest.NewMultiplier(v0_11harvest.Small, 0, sdk.MustNewDecFromStr("0.33")), v0_11harvest.NewMultiplier(v0_11harvest.Medium, 6, sdk.MustNewDecFromStr("0.5")), v0_11harvest.NewMultiplier(v0_11harvest.Large, 24, sdk.OneDec())}),
			v0_11harvest.NewDistributionSchedule(true, "ukava", incentiveGoLiveDate, incentiveEndDate, sdk.NewCoin("hard", sdk.NewInt(63420)), claimEndDate, v0_11harvest.Multipliers{v0_11harvest.NewMultiplier(v0_11harvest.Small, 0, sdk.MustNewDecFromStr("0.33")), v0_11harvest.NewMultiplier(v0_11harvest.Medium, 6, sdk.MustNewDecFromStr("0.5")), v0_11harvest.NewMultiplier(v0_11harvest.Large, 24, sdk.OneDec())}),
		},
		v0_11harvest.DelegatorDistributionSchedules{v0_11harvest.NewDelegatorDistributionSchedule(
			v0_11harvest.NewDistributionSchedule(true, "ukava", incentiveGoLiveDate, incentiveEndDate, sdk.NewCoin("hard", sdk.NewInt(1268392)), claimEndDate, v0_11harvest.Multipliers{v0_11harvest.NewMultiplier(v0_11harvest.Small, 0, sdk.MustNewDecFromStr("0.33")), v0_11harvest.NewMultiplier(v0_11harvest.Medium, 6, sdk.MustNewDecFromStr("0.5")), v0_11harvest.NewMultiplier(v0_11harvest.Large, 24, sdk.OneDec())}),
			time.Hour*24,
		),
		},
	), v0_11harvest.DefaultPreviousBlockTime, v0_11harvest.DefaultDistributionTimes)
	return harvestGS
}
