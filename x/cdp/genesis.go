package cdp

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func InitGenesis(ctx sdk.Context, k Keeper, pk PricefeedKeeper, data GenesisState) {
	// validate denoms - check that any collaterals in the CdpParams are in the pricefeed, pricefeed needs to initgenesis before cdp
	collateralMap := make(map[string]int)
	ap := pk.GetAssetParams(ctx)
	for _, a := range ap.Assets {
		collateralMap[a.AssetCode] = 1
	}

	for _, col := range data.Params.CollateralParams {
		_, found := collateralMap[col.Denom]
		if !found {
			panic(fmt.Sprintf("%s collateral not found in pricefeed", col.Denom))
		}
	}

	k.SetParams(ctx, data.Params)

	for _, cdp := range data.CDPs {
		k.SetCDP(ctx, cdp)
	}

	k.SetGlobalDebt(ctx, data.GlobalDebt)

}

func ExportGenesis(ctx sdk.Context, k Keeper) GenesisState {
	params := k.GetParams(ctx)
	cdps, err := k.GetCDPs(ctx, "", sdk.Dec{})
	if err != nil {
		panic(err)
	}
	debt := k.GetGlobalDebt(ctx)

	return GenesisState{
		Params:     params,
		GlobalDebt: debt,
		CDPs:       cdps,
	}
}
