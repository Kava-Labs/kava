package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/kava/x/evmutil/types"
)

// ConvertCosmosCoinToERC20 locks the initiator's sdk.Coin in the module account
// and mints the receiver a corresponding amount of an ERC20 representing the Coin.
// If a conversion has never been made before and no contract exists, one will be deployed.
// Only denoms registered to the AllowedCosmosDenoms param may be converted.
func (k *Keeper) ConvertCosmosCoinToERC20(
	ctx sdk.Context,
	initiator sdk.AccAddress,
	receiver types.InternalEVMAddress,
	amount sdk.Coin,
) error {
	// check that the conversion is allowed
	tokenInfo, allowed := k.GetAllowedTokenMetadata(ctx, amount.Denom)
	if !allowed {
		return errorsmod.Wrapf(types.ErrSDKConversionNotEnabled, amount.Denom)
	}

	// send coins from initiator to the module account
	// do this before possible contract deploy to prevent unnecessary store interactions
	err := k.bankKeeper.SendCoinsFromAccountToModule(
		ctx, initiator, types.ModuleName, sdk.NewCoins(amount),
	)
	if err != nil {
		return err
	}

	// find deployed contract if it exits
	contractAddress, err := k.GetOrDeployCosmosCoinERC20Contract(ctx, tokenInfo)
	if err != nil {
		return err
	}

	// mint erc20 tokens for the user
	err = k.MintERC20(ctx, contractAddress, receiver, amount.Amount.BigInt())
	if err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeConvertCosmosCoinToERC20,
		sdk.NewAttribute(types.AttributeKeyInitiator, initiator.String()),
		sdk.NewAttribute(types.AttributeKeyReceiver, receiver.String()),
		sdk.NewAttribute(types.AttributeKeyERC20Address, contractAddress.Hex()),
		sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
	))

	return nil
}
