package v0_11

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	v0_11bep3 "github.com/kava-labs/kava/x/bep3/legacy/v0_11"
	v0_9bep3 "github.com/kava-labs/kava/x/bep3/legacy/v0_9"
	v0_11cdp "github.com/kava-labs/kava/x/cdp"
	v0_9cdp "github.com/kava-labs/kava/x/cdp/legacy/v0_9"
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

// MigrateCDP migrates from a v0.9 (or v0.10) cdp genesis state to a v0.11 cdp genesis state
func MigrateCDP(oldGenState v0_9cdp.GenesisState) v0_11cdp.GenesisState {
	var newCDPs v0_11cdp.CDPs
	var newDeposits v0_11cdp.Deposits
	var newCollateralParams v0_11cdp.CollateralParams
	newStartingID := uint64(0)

	for _, cdp := range oldGenState.CDPs {
		newCDP := v0_11cdp.NewCDPWithFees(cdp.ID, cdp.Owner, cdp.Collateral, "bnb-a", cdp.Principal, cdp.AccumulatedFees, cdp.FeesUpdated)
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
		newCollateralParam := v0_11cdp.NewCollateralParam(cp.Denom, "bnb-a", cp.LiquidationRatio, cp.DebtLimit, cp.StabilityFee, cp.AuctionSize, cp.LiquidationPenalty, 0x01, cp.SpotMarketID, cp.LiquidationMarketID, cp.ConversionFactor)
		newCollateralParams = append(newCollateralParams, newCollateralParam)
	}
	btcbCollateralParam := v0_11cdp.NewCollateralParam("btcb", "btcb-a", sdk.MustNewDecFromStr("1.5"), sdk.NewCoin("usdx", sdk.NewInt(100000000000)), sdk.MustNewDecFromStr("1.000000001547125958"), sdk.NewInt(100000000), sdk.MustNewDecFromStr("0.075000000000000000"), 0x02, "btc:usd", "btc:usd:30", sdk.NewInt(8))
	busdaCollateralParam := v0_11cdp.NewCollateralParam("busd", "busd-a", sdk.MustNewDecFromStr("1.01"), sdk.NewCoin("usdx", sdk.NewInt(3000000000000)), sdk.OneDec(), sdk.NewInt(1000000000000), sdk.MustNewDecFromStr("0.075000000000000000"), 0x03, "busd:usd", "busd:usd:30", sdk.NewInt(8))
	busdbCollateralParam := v0_11cdp.NewCollateralParam("busd", "busd-b", sdk.MustNewDecFromStr("1.1"), sdk.NewCoin("usdx", sdk.NewInt(1000000000000)), sdk.MustNewDecFromStr("1.000000012857214317"), sdk.NewInt(1000000000000), sdk.MustNewDecFromStr("0.075000000000000000"), 0x04, "busd:usd", "busd:usd:30", sdk.NewInt(8))
	xrpbCollateralParam := v0_11cdp.NewCollateralParam("xrpb", "xrpb-a", sdk.MustNewDecFromStr("1.5"), sdk.NewCoin("usdx", sdk.NewInt(100000000000)), sdk.MustNewDecFromStr("1.000000001547125958"), sdk.NewInt(100000000000), sdk.MustNewDecFromStr("0.075000000000000000"), 0x05, "xrp:usd", "xrp:usd:30", sdk.NewInt(8))
	newCollateralParams = append(newCollateralParams, btcbCollateralParam, busdaCollateralParam, busdbCollateralParam, xrpbCollateralParam)
	oldDebtParam := oldGenState.Params.DebtParam

	newDebtParam := v0_11cdp.NewDebtParam(oldDebtParam.Denom, oldDebtParam.ReferenceAsset, oldDebtParam.ConversionFactor, oldDebtParam.DebtFloor, oldDebtParam.SavingsRate)

	newGlobalDebtLimit := oldGenState.Params.GlobalDebtLimit.Add(btcbCollateralParam.DebtLimit).Add(busdaCollateralParam.DebtLimit).Add(busdbCollateralParam.DebtLimit).Add(xrpbCollateralParam.DebtLimit)

	newParams := v0_11cdp.NewParams(newGlobalDebtLimit, newCollateralParams, newDebtParam, oldGenState.Params.SurplusAuctionThreshold, oldGenState.Params.SurplusAuctionLot, oldGenState.Params.DebtAuctionThreshold, oldGenState.Params.DebtAuctionLot, oldGenState.Params.SavingsDistributionFrequency, false)

	return v0_11cdp.NewGenesisState(
		newParams,
		newCDPs,
		newDeposits,
		newStartingID,
		oldGenState.DebtDenom,
		oldGenState.GovDenom,
		oldGenState.PreviousDistributionTime,
		sdk.ZeroInt(),
	)

}
