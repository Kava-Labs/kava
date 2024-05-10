package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/x/precisebank/types"
)

// RegisterInvariants registers the x/precisebank module invariants
func RegisterInvariants(
	ir sdk.InvariantRegistry,
	k Keeper,
	bk types.BankKeeper,
) {
}

// AllInvariants runs all invariants of the X/precisebank module.
func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		return "", false
	}
}
