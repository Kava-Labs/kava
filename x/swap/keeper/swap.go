package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	//sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	//"github.com/kava-labs/kava/x/swap/types"
)

// SwapExactForTokens swaps an exact coin a input for a coin b output
func (k *Keeper) SwapExactForTokens(ctx sdk.Context, requester sdk.AccAddress, exactCoinA, coinB sdk.Coin, slippage sdk.Dec) error {
	return nil
}

// SwapExactForTokens swaps a coin a input for an exact coin b output
func (k *Keeper) SwapForExactTokens(ctx sdk.Context, requester sdk.AccAddress, exactCoinA, coinB sdk.Coin, slippage sdk.Dec) error {
	return nil
}
