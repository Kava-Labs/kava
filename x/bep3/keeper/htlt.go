package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/bep3/types"
)

// AddHTLT adds an htlt
func (k Keeper) AddHTLT(ctx sdk.Context, from sdk.AccAddress, to sdk.AccAddress, recipientOtherChain,
	senderOtherChain string, randomNumberHash string, timestamp int64, asset sdk.Coins,
	expectedIncome string, heightSpan int64, crossChain bool) (string, sdk.Error) {

	err := k.ValidateAsset(ctx, asset)
	if err != nil {
		return "", err
	}

	htlt := types.NewHTLT(from, to, recipientOtherChain, senderOtherChain,
		randomNumberHash, timestamp, asset, expectedIncome, heightSpan,
		crossChain)

	swapID, sdkErr := k.StoreNewHTLT(ctx, htlt)
	if sdkErr != nil {
		return "", sdk.ErrInternal(sdkErr.Error())
	}
	// Emit event 'htlt_created'
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCreateHtlt,
			sdk.NewAttribute(types.AttributeKeyHtltSwapID, fmt.Sprintf("%s", swapID)),
			sdk.NewAttribute(types.AttributeKeyFrom, fmt.Sprintf("%s", htlt.From)),
			sdk.NewAttribute(types.AttributeKeyTo, fmt.Sprintf("%s", htlt.To)),
			sdk.NewAttribute(types.AttributeKeyCoinDenom, fmt.Sprintf("%s", htlt.Amount[0].Denom)),
			sdk.NewAttribute(types.AttributeKeyCoinAmount, fmt.Sprintf("%d", htlt.Amount[0].Amount.Int64())),
		),
	)

	return swapID, nil
}

// GetAllHtlts returns all HTLTs from the store
func (k Keeper) GetAllHtlts(ctx sdk.Context) (htlts types.HTLTs) {
	k.IterateHTLTs(ctx, func(htlt types.HTLT) bool {
		htlts = append(htlts, htlt)
		return false
	})
	return
}

// ValidateAsset validates that a amount is valid for HTLTs
func (k Keeper) ValidateAsset(ctx sdk.Context, assets sdk.Coins) sdk.Error {
	if len(assets) != 1 {
		return sdk.ErrInternal("HTLTs currently only support 1 asset at a time")
	}
	// TODO: Validate that this asset is in module params
	return nil
}
