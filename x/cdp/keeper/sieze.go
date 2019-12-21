package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/cdp/types"
	liqtypes "github.com/kava-labs/kava/x/liquidator/types"
)

// SeizeCollateral sends collateral for all deposits to the liquidator module and moves cdp debt to the liquidator module
func (k Keeper) SeizeCollateral(ctx sdk.Context, cdp types.CDP) {
	deposits := k.GetDeposits(ctx, cdp.ID)
	for _, dep := range deposits {
		if !dep.InLiquidation {
			dep.InLiquidation = true
			k.SetDeposit(ctx, dep, cdp.ID)
			err := k.supplyKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, liqtypes.ModuleName, dep.Amount)
			if err != nil {
				panic(err)
			}
		}
	}
	debtAmt := sdk.ZeroInt()
	for _, dc := range cdp.Principal {
		debtAmt = debtAmt.Add(dc.Amount)
	}
	for _, dc := range cdp.AccumulatedFees {
		debtAmt = debtAmt.Add(dc.Amount)
	}
	debtCoins := sdk.NewCoins(sdk.NewCoin(k.GetDebtDenom(ctx), debtAmt))
	err := k.supplyKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, liqtypes.ModuleName, debtCoins)
	if err != nil {
		panic(err)
	}
}
