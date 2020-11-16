package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/harvest/types"
)

// ApplyInterestRateUpdates translates the current interest rate models from the params to the store
func (k Keeper) ApplyInterestRateUpdates(ctx sdk.Context) {
	denomSet := map[string]bool{}

	params := k.GetParams(ctx)
	for _, mm := range params.MoneyMarkets {
		model, found := k.GetInterestRateModel(ctx, mm.Denom)
		if !found {
			k.SetInterestRateModel(ctx, mm.Denom, mm.InterestRateModel)
			continue
		}
		if !model.Equal(mm.InterestRateModel) {
			// TODO: call AccrueInterest for the asset type here (for all addresses?)
			k.SetInterestRateModel(ctx, mm.Denom, mm.InterestRateModel)
		}
		denomSet[mm.Denom] = true
	}

	k.IterateInterestRateModels(ctx, func(denom string, i types.InterestRateModel) bool {
		if !denomSet[denom] {
			k.DeleteInterestRateModel(ctx, denom)
		}
		return false
	})
}

// AccrueInterest
func (k Keeper) AccrueInterest(ctx sdk.Context, denom string) error {
	return nil
}
