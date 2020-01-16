package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/pricefeed/types"
)

// GetParams gets params from the store
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	return types.NewParams(k.GetMarketParams(ctx))
}

// SetParams updates params in the store
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}

// GetMarketParams get asset params from store
func (k Keeper) GetMarketParams(ctx sdk.Context) types.Markets {
	var markets types.Markets
	k.paramstore.Get(ctx, types.KeyMarkets, &markets)
	return markets
}

// GetOracles returns the oracles in the pricefeed store
func (k Keeper) GetOracles(ctx sdk.Context, marketID string) ([]sdk.AccAddress, sdk.Error) {

	for _, m := range k.GetMarketParams(ctx) {
		if marketID == m.MarketID {
			return m.Oracles, nil
		}
	}
	return []sdk.AccAddress{}, types.ErrInvalidMarket(k.Codespace(), marketID)
}

// GetOracle returns the oracle from the store or an error if not found
func (k Keeper) GetOracle(ctx sdk.Context, marketID string, address sdk.AccAddress) (sdk.AccAddress, sdk.Error) {
	oracles, err := k.GetOracles(ctx, marketID)
	if err != nil {
		return sdk.AccAddress{}, types.ErrInvalidMarket(k.Codespace(), marketID)
	}
	for _, addr := range oracles {
		if address.Equals(addr) {
			return addr, nil
		}
	}
	return sdk.AccAddress{}, types.ErrInvalidOracle(k.codespace, address)
}

// GetMarket returns the market if it is in the pricefeed system
func (k Keeper) GetMarket(ctx sdk.Context, marketID string) (types.Market, bool) {
	markets := k.GetMarketParams(ctx)

	for i := range markets {
		if markets[i].MarketID == marketID {
			return markets[i], true
		}
	}
	return types.Market{}, false

}
