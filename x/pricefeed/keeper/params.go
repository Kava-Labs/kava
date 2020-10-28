package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/pricefeed/types"
)

// GetParams returns the params from the store
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	var p types.Params
	k.paramSubspace.GetParamSet(ctx, &p)
	return p
}

// SetParams sets params on the store
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSubspace.SetParamSet(ctx, &params)
}

// GetMarkets returns the markets from params
func (k Keeper) GetMarkets(ctx sdk.Context) types.Markets {
	return k.GetParams(ctx).Markets
}

// GetOracles returns the oracles in the pricefeed store
func (k Keeper) GetOracles(ctx sdk.Context, marketID string) ([]sdk.AccAddress, error) {
	for _, m := range k.GetMarkets(ctx) {
		if marketID == m.MarketID {
			return m.Oracles, nil
		}
	}
	return []sdk.AccAddress{}, sdkerrors.Wrap(types.ErrInvalidMarket, marketID)
}

// GetOracle returns the oracle from the store or an error if not found
func (k Keeper) GetOracle(ctx sdk.Context, marketID string, address sdk.AccAddress) (sdk.AccAddress, error) {
	oracles, err := k.GetOracles(ctx, marketID)
	if err != nil {
		return sdk.AccAddress{}, sdkerrors.Wrap(types.ErrInvalidMarket, marketID)
	}
	for _, addr := range oracles {
		if address.Equals(addr) {
			return addr, nil
		}
	}
	return sdk.AccAddress{}, sdkerrors.Wrap(types.ErrInvalidOracle, address.String())
}

// GetMarket returns the market if it is in the pricefeed system
func (k Keeper) GetMarket(ctx sdk.Context, marketID string) (types.Market, bool) {
	markets := k.GetMarkets(ctx)

	for i := range markets {
		if markets[i].MarketID == marketID {
			return markets[i], true
		}
	}
	return types.Market{}, false
}
