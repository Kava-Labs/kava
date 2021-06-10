package keeper

import (
	"github.com/kava-labs/kava/x/swap/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) GetShares(depositor sdk.AccAddress, pool types.Pool) (sdk.Int, bool) {
	return sdk.ZeroInt(), false
}
