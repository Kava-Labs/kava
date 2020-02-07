package keeper

import (
	"fmt"

	binance "github.com/binance-chain/go-sdk/common/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/bep3/types"
)

// AddHTLT adds an htlt
func (k Keeper) AddHTLT(ctx sdk.Context, from sdk.AccAddress, to sdk.AccAddress, recipientOtherChain,
	senderOtherChain string, randomNumberHash binance.SwapBytes, timestamp int64, amount binance.Coins,
	expectedIncome string, heightSpan int64, crossChain bool) ([]byte, sdk.Error) {

	// validation
	err := k.ValidateAsset(ctx, amount)
	if err != nil {
		return []byte{}, err
	}

	// Create new HTLT
	htlt := types.NewHTLT(from, to, recipientOtherChain, senderOtherChain,
		randomNumberHash, timestamp, amount, expectedIncome, heightSpan,
		crossChain)

	swapID := k.StoreNewHTLT(ctx, htlt)

	// Emit event 'htlt_created'
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCreateHtlt,
			sdk.NewAttribute(types.AttributeKeyHtltSwapID, fmt.Sprintf("%d", swapID)),
			sdk.NewAttribute(types.AttributeKeyFrom, fmt.Sprintf("%s", htlt.From)),
			sdk.NewAttribute(types.AttributeKeyTo, fmt.Sprintf("%s", htlt.To)),
			sdk.NewAttribute(types.AttributeKeyCoinDenom, fmt.Sprintf("%s", htlt.Amount[0].Denom)),
			sdk.NewAttribute(types.AttributeKeyCoinAmount, fmt.Sprintf("%d", htlt.Amount[0].Amount)),
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
func (k Keeper) ValidateAsset(ctx sdk.Context, assets binance.Coins) sdk.Error {
	if len(assets) != 1 {
		return sdk.ErrInternal("HTLTs currently only support 1 asset at a time")
	}
	// _, found := k.GetAsset(ctx, amount[0].Denom)
	// if !found {
	// return types.ErrCollateralNotSupported(k.codespace, collateral[0].Denom)
	// }
	return nil
}
