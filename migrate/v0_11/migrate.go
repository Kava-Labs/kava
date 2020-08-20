package v0_11

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	v0_11bep3 "github.com/kava-labs/kava/x/bep3/legacy/v0_11"
	v0_9bep3 "github.com/kava-labs/kava/x/bep3/legacy/v0_9"
	v0_11cdp "github.com/kava-labs/kava/x/cdp/legacy/v0_11"
	v0_9cdp "github.com/kava-labs/kava/x/cdp/legacy/v0_9"
)

// MigrateBep3 migrates from a v0.9 (or v0.10) bep3 genesis state to a v0.11 bep3 genesis state
func MigrateBep3(oldGenState v0_9bep3.GenesisState) v0_11bep3.GenesisState {
	var assetParams v0_11bep3.AssetParams
	v0_9Params := oldGenState.Params

	for _, asset := range v0_9Params.SupportedAssets {
		v10AssetParam := v0_11bep3.AssetParam{
			Active:        asset.Active,
			Denom:         asset.Denom,
			CoinID:        asset.CoinID,
			DeputyAddress: v0_9Params.BnbDeputyAddress,
			FixedFee:      v0_9Params.BnbDeputyFixedFee,
			MinSwapAmount: v0_9Params.MinAmount,
			MaxSwapAmount: v0_9Params.MaxAmount,
			MinBlockLock:  v0_9Params.MinBlockLock,
			MaxBlockLock:  v0_9Params.MaxBlockLock,
			SupplyLimit: v0_11bep3.AssetSupply{
				SupplyLimit:    sdk.NewCoin(asset.Denom, sdk.ZeroInt()),
				CurrentSupply:  sdk.NewCoin(asset.Denom, sdk.ZeroInt()),
				IncomingSupply: sdk.NewCoin(asset.Denom, sdk.ZeroInt()),
				OutgoingSupply: sdk.NewCoin(asset.Denom, sdk.ZeroInt()),
			},
		}
		assetParams = append(assetParams, v10AssetParam)
	}
	for _, supply := range oldGenState.AssetSupplies {
		for _, asset := range assetParams {
			if asset.Denom == supply.Denom {
				asset.SupplyLimit.SupplyLimit = supply.SupplyLimit
				asset.SupplyLimit.CurrentSupply = supply.CurrentSupply
				asset.SupplyLimit.IncomingSupply = supply.IncomingSupply
				asset.SupplyLimit.OutgoingSupply = supply.OutgoingSupply
			}
		}
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
		AtomicSwaps: swaps,
	}
}

// MigrateCDP migrates from a v0.9 (or v0.10) cdp genesis state to a v0.11 cdp genesis state
func MigrateCDP(oldGenState v0_9cdp.GenesisState) v0_11cdp.GenesisState {
	var newCDPs v0_11cdp.CDPs
	var newDeposits v0_11cdp.Deposits
	var newCollateralParams v0_11cdp.CollateralParams
	newStartingID := uint64(0)

	for _, cdp := range oldGenState.CDPs {
		newCDP := v0_11cdp.NewCDP(cdp.ID, cdp.Owner, cdp.Collateral, "bnb-a", cdp.Principal, cdp.AccumulatedFees, cdp.FeesUpdated)
		newCDPs = append(newCDPs, newCDP)
		if cdp.ID >= newStartingID {
			newStartingID = cdp.ID + 1
		}
	}

	for _, dep := range oldGenState.Deposits {
		newDep := v0_11cdp.NewDeposit(dep.CdpID, dep.Depositor, dep.Amount)
		newDeposits = append(newDeposits, newDep)
	}

	for _, cp := range oldGenState.Params.CollateralParams {
		newCollateralParam := v0_11cdp.NewCollateralParam(cp.Denom, "bnb-a", cp.LiquidationRatio, cp.DebtLimit, cp.StabilityFee, cp.AuctionSize, cp.LiquidationPenalty, cp.Prefix, cp.SpotMarketID, cp.LiquidationMarketID, cp.ConversionFactor)
		newCollateralParams = append(newCollateralParams, newCollateralParam)
	}

	oldDebtParam := oldGenState.Params.DebtParam

	newDebtParam := v0_11cdp.NewDebtParam(oldDebtParam.Denom, oldDebtParam.ReferenceAsset, oldDebtParam.ConversionFactor, oldDebtParam.DebtFloor, oldDebtParam.SavingsRate)

	newParams := v0_11cdp.NewParams(oldGenState.Params.GlobalDebtLimit, newCollateralParams, newDebtParam, oldGenState.Params.SurplusAuctionThreshold, oldGenState.Params.SurplusAuctionLot, oldGenState.Params.DebtAuctionThreshold, oldGenState.Params.DebtAuctionLot, oldGenState.Params.SavingsDistributionFrequency, false)

	return v0_11cdp.GenesisState{
		Params:                   newParams,
		CDPs:                     newCDPs,
		Deposits:                 newDeposits,
		StartingCdpID:            newStartingID,
		DebtDenom:                oldGenState.DebtDenom,
		GovDenom:                 oldGenState.GovDenom,
		PreviousDistributionTime: oldGenState.PreviousDistributionTime,
	}

}
