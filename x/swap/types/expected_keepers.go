package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SwapHooks are event hooks called when a user's deposit to a swap pool changes.
type SwapHooks interface {
	AfterPoolDepositCreated(ctx sdk.Context, poolID string, depositor sdk.AccAddress, sharedOwned sdk.Int)
	BeforePoolDepositModified(ctx sdk.Context, poolID string, depositor sdk.AccAddress, sharedOwned sdk.Int)
}
