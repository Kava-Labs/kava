package keeper

import (
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ApplyInterestRateUpdates translates the current interest rate models from the params to the store
func (k Keeper) ApplyInterestRateUpdates(ctx sdk.Context) {
	params := k.GetParams(ctx)
	for _, mm := range params.MoneyMarkets {
		model, found := k.GetInterestRateModel(ctx, mm.Denom)
		if !found {
			k.SetInterestRateModel(ctx, mm.Denom, mm.InterestRateModel)
			continue
		}
		if !reflect.DeepEqual(model, mm.InterestRateModel) {
			k.SetInterestRateModel(ctx, mm.Denom, mm.InterestRateModel)
		}
	}
}
