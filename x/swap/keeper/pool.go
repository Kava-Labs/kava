package keeper

import (
	"github.com/kava-labs/kava/x/swap/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) Deposit(ctx sdk.Context, depositor sdk.AccAddress, amountA sdk.Coin, amountB sdk.Coin) error {
	return nil
}

func (k Keeper) GetPool(poolName string) (types.Pool, bool) {
	return types.Pool{}, false
}
