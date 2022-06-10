package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kava-labs/kava/x/liquidstaking/types"
)

// MintDerivative mints a new derivative
func (k Keeper) MintDerivative(ctx sdk.Context, validator sdk.ValAddress, coin sdk.Coin) error {
	currDerivative, foundDerivative := k.GetDerivative(ctx, validator)

	derivative := types.NewDerivative(validator, coin)
	if foundDerivative {
		derivative.Amount = derivative.Amount.Add(currDerivative.Amount)

	}

	k.SetDerivative(ctx, derivative)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMintDerivative,
			sdk.NewAttribute(sdk.AttributeKeyAmount, coin.String()),
			sdk.NewAttribute(types.AttributeKeyValidator, derivative.Validator.String()),
		),
	)

	return nil
}

// BurnDerivative burns an existing derivative
func (k Keeper) BurnDerivative(ctx sdk.Context, validator sdk.ValAddress, coin sdk.Coin) error {

	currDerivative, foundDerivative := k.GetDerivative(ctx, validator)
	if !foundDerivative {
		return sdkerrors.Wrap(types.ErrNoDerivativeFound, fmt.Sprintf(" for address: %s", validator.String()))
	}

	if coin.Amount.GT(currDerivative.Amount.Amount) {
		return sdkerrors.Wrap(types.ErrInvalidBurnAmount, fmt.Sprintf("%s > %s", coin.Amount.String(), currDerivative.Amount.String()))
	}

	derivative := types.NewDerivative(validator, coin)
	if foundDerivative {
		derivative.Amount = derivative.Amount.Sub(currDerivative.Amount)
		// TODO: call hook if needed

	}

	k.SetDerivative(ctx, derivative)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMintDerivative,
			sdk.NewAttribute(sdk.AttributeKeyAmount, coin.String()),
			sdk.NewAttribute(types.AttributeKeyValidator, derivative.Validator.String()),
		),
	)

	return nil
}
