package hard_supply

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/incentive/types"
)

var _ types.SourceAdapter = SourceAdapter{}

type SourceAdapter struct {
	keeper types.HardKeeper
}

func NewSourceAdapter(keeper types.HardKeeper) SourceAdapter {
	return SourceAdapter{
		keeper: keeper,
	}
}

func (f SourceAdapter) TotalSharesBySource(ctx sdk.Context, sourceID string) sdk.Dec {
	coins, found := f.keeper.GetSuppliedCoins(ctx)
	if !found {
		return sdk.ZeroDec()
	}

	totalSupplied := coins.AmountOf(sourceID).ToDec()

	interestFactor, found := f.keeper.GetSupplyInterestFactor(ctx, sourceID)
	if !found {
		// assume nothing has been borrowed so the factor starts at it's default value
		interestFactor = sdk.OneDec()
	}

	// return supplied/factor to get the "pre interest" value of the current total supplied
	return totalSupplied.Quo(interestFactor)
}

func (f SourceAdapter) OwnerSharesBySource(
	ctx sdk.Context,
	owner sdk.AccAddress,
	sourceIDs []string,
) map[string]sdk.Dec {
	depositCoins := sdk.NewDecCoins()

	deposit, found := f.keeper.GetDeposit(ctx, owner)
	if found {
		normalizedDeposit, err := deposit.NormalizedDeposit()
		if err != nil {
			panic(fmt.Errorf("failed to normalize hard deposit for owner %s: %w", owner, err))
		}

		depositCoins = normalizedDeposit
	}

	shares := make(map[string]sdk.Dec)
	for _, id := range sourceIDs {
		shares[id] = depositCoins.AmountOf(id)
	}

	return shares
}
