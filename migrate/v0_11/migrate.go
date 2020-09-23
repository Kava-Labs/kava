package v0_11

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	v0_11bep3 "github.com/kava-labs/kava/x/bep3"
	v0_9bep3 "github.com/kava-labs/kava/x/bep3/legacy/v0_9"
)

// MigrateBep3 migrates from a v0.9 (or v0.10) bep3 genesis state to a v0.11 bep3 genesis state
func MigrateBep3(oldGenState v0_9bep3.GenesisState) v0_11bep3.GenesisState {
	var assetParams v0_11bep3.AssetParams
	var assetSupplies v0_11bep3.AssetSupplies
	v0_9Params := oldGenState.Params

	for _, asset := range v0_9Params.SupportedAssets {
		v11AssetParam := v0_11bep3.AssetParam{
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
		assetParams = append(assetParams, v11AssetParam)
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

	// -------------- ADD BTCB To BEP3 params --------------------
	btcbAssetParam := v0_11bep3.NewAssetParam(
		"btcb",
		0,
		v0_11bep3.SupplyLimit{
			Limit:          sdk.NewInt(10000000000), // 100 BTC limit at launch
			TimeLimited:    false,
			TimePeriod:     time.Duration(0),
			TimeBasedLimit: sdk.ZeroInt()},
		true,
		v0_9Params.BnbDeputyAddress, // TODO get an additional deputy address from binance
		v0_9Params.BnbDeputyFixedFee,
		sdk.OneInt(),
		sdk.NewInt(1000000000),
		220,
		270,
	)
	btcbAssetSupply := v0_11bep3.NewAssetSupply(
		sdk.NewCoin("btcb", sdk.ZeroInt()),
		sdk.NewCoin("btcb", sdk.ZeroInt()),
		sdk.NewCoin("btcb", sdk.ZeroInt()),
		sdk.NewCoin("btcb", sdk.ZeroInt()),
		time.Duration(0))
	assetParams = append(assetParams, btcbAssetParam)
	assetSupplies = append(assetSupplies, btcbAssetSupply)
	xrpbAssetParam := v0_11bep3.NewAssetParam(
		"xrpb",
		144,
		v0_11bep3.SupplyLimit{
			Limit:          sdk.NewInt(1000000000000), // 1,000,000 XRP limit at launch
			TimeLimited:    false,
			TimePeriod:     time.Duration(0),
			TimeBasedLimit: sdk.ZeroInt()},
		true,
		v0_9Params.BnbDeputyAddress, // TODO  get an additional deputy address from binance
		v0_9Params.BnbDeputyFixedFee,
		sdk.OneInt(),
		sdk.NewInt(100000000000),
		220,
		270,
	)
	xrpbAssetSupply := v0_11bep3.NewAssetSupply(
		sdk.NewCoin("xrpb", sdk.ZeroInt()),
		sdk.NewCoin("xrpb", sdk.ZeroInt()),
		sdk.NewCoin("xrpb", sdk.ZeroInt()),
		sdk.NewCoin("xrpb", sdk.ZeroInt()),
		time.Duration(0))
	assetParams = append(assetParams, xrpbAssetParam)
	assetSupplies = append(assetSupplies, xrpbAssetSupply)
	busdAssetParam := v0_11bep3.NewAssetParam(
		"busd",
		727, // note - no official SLIP 44 ID
		v0_11bep3.SupplyLimit{
			Limit:          sdk.NewInt(10000000000000), // 100,000 BUSD limit at launch
			TimeLimited:    false,
			TimePeriod:     time.Duration(0),
			TimeBasedLimit: sdk.ZeroInt()},
		true,
		v0_9Params.BnbDeputyAddress, // TODO  get an additional deputy address from binance
		v0_9Params.BnbDeputyFixedFee,
		sdk.OneInt(),
		sdk.NewInt(1000000000000),
		220,
		270,
	)
	busdAssetSupply := v0_11bep3.NewAssetSupply(
		sdk.NewCoin("busd", sdk.ZeroInt()),
		sdk.NewCoin("busd", sdk.ZeroInt()),
		sdk.NewCoin("busd", sdk.ZeroInt()),
		sdk.NewCoin("busd", sdk.ZeroInt()),
		time.Duration(0))
	assetParams = append(assetParams, busdAssetParam)
	assetSupplies = append(assetSupplies, busdAssetSupply)
	return v0_11bep3.GenesisState{
		Params:            v0_11bep3.NewParams(assetParams),
		AtomicSwaps:       swaps,
		Supplies:          assetSupplies,
		PreviousBlockTime: v0_11bep3.DefaultPreviousBlockTime,
	}
}
