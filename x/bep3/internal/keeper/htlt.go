package keeper

import (
	"fmt"

	bnb "github.com/binance-chain/go-sdk/common/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Kava-Labs/kava/x/bep3/internal/types"
)

// AddHtlt adds an htlt
func (k Keeper) AddHtlt(ctx sdk.Context, from bnb.AccAddress, to bnb.AccAddress, recipientOtherChain,
	senderOtherChain string, randomNumberHash bnb.SwapBytes, timestamp int64, amount bnb.Coins,
	expectedIncome string, heightSpan int64, crossChain bool) (uint64, sdk.Error) {

	// validation
	err := k.ValidateAmount(ctx, amount)
	if err != nil {
		return 0, err
	}
	// _, found := k.GetCdpByOwnerAndDenom(ctx, owner, collateral[0].Denom)
	// if found {
	// 	return types.ErrCdpAlreadyExists(k.codespace, owner, collateral[0].Denom)
	// }
	// err = k.ValidatePrincipalAdd(ctx, principal)
	// if err != nil {
	// 	return err
	// }

	// Create new KHTLT from HTLT
	htlt := types.NewHTLT(from, to, recipientOtherChain, senderOtherChain,
		randomNumberHash, timestamp, amount, expectedIncome, heightSpan,
		crossChain)

	khtlt := types.NewKHTLT(htlt)

	// TODO: This assumes [CrossChain false = Msg originated on Kava]
	if !crossChain {
		// TODO: Validate that the address is good

		// Parse [from, amount] from Binance types -> Cosmos types
		sdkFrom, err := sdk.AccAddressFromBech32(from.String())
		if err != nil {
			return 0, sdk.ErrInvalidAddress(fmt.Sprintf("%s", err))
		}
		sdkAmount := sdk.NewCoins(sdk.NewInt64Coin(amount[0].Denom, amount[0].Amount))

		// Send coins from the address on chain t
		err = k.supplyKeeper.SendCoinsFromAccountToModule(ctx, sdkFrom, types.ModuleName, sdkAmount)
		if err != nil {
			return 0, sdk.ErrInternal(fmt.Sprintf("%s", err))
		}
	}

	khtltID, err := k.StoreNewKHTLT(ctx, khtlt)
	if err != nil {
		return 0, err
	}
	// TODO: k.IncrementTotalLocked(ctx, amount[0].Denom, amount[0].Amount)
	//		from inside k.StoreNewKHTLT

	// Emit event 'htlt_created'
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCreateHtlt,
			sdk.NewAttribute(types.AttributeKeyHtltID, fmt.Sprintf("%d", khtltID)),
			sdk.NewAttribute(types.AttributeKeyFrom, fmt.Sprintf("%s", khtlt.HTLT.From)),
			sdk.NewAttribute(types.AttributeKeyTo, fmt.Sprintf("%s", khtlt.HTLT.To)),
			sdk.NewAttribute(types.AttributeKeyCoinDenom, fmt.Sprintf("%s", khtlt.HTLT.Amount[0].Denom)),
			sdk.NewAttribute(types.AttributeKeyCoinAmount, fmt.Sprintf("%d", khtlt.HTLT.Amount[0].Amount)),
		),
	)

	return khtltID, nil
}

// ValidateAmount validates that a amount is valid for HTLTs
func (k Keeper) ValidateAmount(ctx sdk.Context, amount bnb.Coins) sdk.Error {
	if len(amount) != 1 {
		// return types.ErrInvalidCollateralLength(k.codespace, len(collateral))
	}
	// _, found := k.GetAsset(ctx, amount[0].Denom)
	// if !found {
	// return types.ErrCollateralNotSupported(k.codespace, collateral[0].Denom)
	// }
	return nil
}
