package keeper

import (
	"bytes"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/evmutil/types"
)

// GetParams returns the total set of evm parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSubspace.GetParamSetIfExists(ctx, &params)
	return params
}

// SetParams sets the evm parameters to the param space.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSubspace.SetParamSet(ctx, &params)
}

// GetAllowedTokenMetadata gets the token metadata for the given cosmosDenom if it is allowed.
// Returns the metadata if allowed, and a bool indicating if the denom was in the allow list or not.
func (k Keeper) GetAllowedTokenMetadata(ctx sdk.Context, cosmosDenom string) (types.AllowedCosmosCoinERC20Token, bool) {
	params := k.GetParams(ctx)
	for _, token := range params.AllowedCosmosDenoms {
		if token.CosmosDenom == cosmosDenom {
			return token, true
		}
	}
	return types.AllowedCosmosCoinERC20Token{}, false
}

// GetEnabledConversionPairFromERC20Address returns an ConversionPair from the internal contract address.
func (k Keeper) GetEnabledConversionPairFromERC20Address(
	ctx sdk.Context,
	address types.InternalEVMAddress,
) (types.ConversionPair, error) {
	params := k.GetParams(ctx)
	for _, pair := range params.EnabledConversionPairs {
		if bytes.Equal(pair.KavaERC20Address, address.Bytes()) {
			return pair, nil
		}
	}

	return types.ConversionPair{}, errorsmod.Wrap(types.ErrConversionNotEnabled, address.String())
}

// GetEnabledConversionPairFromDenom returns an ConversionPair from the sdk.Coin denom.
func (k Keeper) GetEnabledConversionPairFromDenom(
	ctx sdk.Context,
	denom string,
) (types.ConversionPair, error) {
	params := k.GetParams(ctx)
	for _, pair := range params.EnabledConversionPairs {
		if pair.Denom == denom {
			return pair, nil
		}
	}

	return types.ConversionPair{}, errorsmod.Wrap(types.ErrConversionNotEnabled, denom)
}
