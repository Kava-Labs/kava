package v0_10

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	v0_10bep3 "github.com/kava-labs/kava/x/bep3/legacy/v0_10"
	v0_9bep3 "github.com/kava-labs/kava/x/bep3/legacy/v0_9"
)

func MigrateBep3(oldGenState v0_9bep3.GenesisState) v0_10bep3.GenesisState {
	var assetParams v0_10bep3.AssetParams
	v0_9Params := oldGenState.Params

	for _, asset := range v0_9Params.SupportedAssets {
		v10AssetParam := v0_10bep3.AssetParam{
			Active:               asset.Active,
			Denom:                asset.Denom,
			CoinID:               asset.CoinID,
			DeputyAddress:        v0_9Params.BnbDeputyAddress,
			IncomingSwapFixedFee: v0_9Params.BnbDeputyFixedFee,
			MinSwapAmount:        v0_9Params.MinAmount,
			MaxSwapAmount:        v0_9Params.MaxAmount,
			MinBlockLock:         v0_9Params.MinBlockLock,
			MaxBlockLock:         v0_9Params.MaxBlockLock,
			SupplyLimit: v0_10bep3.AssetSupply{
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
	var swaps v0_10bep3.AtomicSwaps
	for _, oldSwap := range oldGenState.AtomicSwaps {
		newSwap := v0_10bep3.AtomicSwap{
			Amount:              oldSwap.Amount,
			RandomNumberHash:    oldSwap.RandomNumberHash,
			ExpireHeight:        oldSwap.ExpireHeight,
			Timestamp:           oldSwap.Timestamp,
			Sender:              oldSwap.Sender,
			Recipient:           oldSwap.Recipient,
			SenderOtherChain:    oldSwap.SenderOtherChain,
			RecipientOtherChain: oldSwap.RecipientOtherChain,
			ClosedBlock:         oldSwap.ClosedBlock,
			Status:              v0_10bep3.SwapStatus(oldSwap.Status),
			CrossChain:          oldSwap.CrossChain,
			Direction:           v0_10bep3.SwapDirection(oldSwap.Direction),
		}
		swaps = append(swaps, newSwap)
	}
	return v0_10bep3.GenesisState{
		Params: v0_10bep3.Params{
			AssetParams: assetParams},
		AtomicSwaps: swaps,
	}
}
